package utilite

import (
	git "github.com/libgit2/git2go/v30"
)

func copyMap(m map[string]bool) map[string]bool {
	nm := make(map[string]bool)
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

func GetCommit(hash string, repo *git.Repository) (*git.Commit, error) {
	obj, err := repo.RevparseSingle(hash)
	if err != nil {
		return nil, err
	}
	commit, err := obj.AsCommit()
	return commit, err
}

func getParents(commit *git.Commit) []*git.Commit {
	parents := make([]*git.Commit, commit.ParentCount())
	for i := 0; i < len(parents); i++ {
		parents[i] = commit.Parent(uint(i))
	}
	return parents
}

// updateCommit recursively updates
// all commits dependent of renamed ones.
// It does this by doing DFS.
func updateCommit(
	commit *git.Commit,
	places map[string]bool,
	visited map[string]*git.Commit,
	counter int,
	newMsg map[string]string,
	repo *git.Repository,
) (*git.Commit, error) {
	// Check if current commit was already updated
	if newCommit, ok := visited[commit.Id().String()]; ok {
		return newCommit, nil
	}

	// Check if current commit is to be renamed
	cid := commit.Id().String()
	if visited, ok := places[cid]; ok && !visited {
		places[cid] = true
		counter += 1
	}

	changed := false

	// Get message and update if needed
	message := commit.Message()
	if newMsg, ok := newMsg[commit.Id().String()]; ok {
		message = newMsg
		changed = true
	}

	parents := getParents(commit)
	// if there are still commits to be visited then do recursion
	// and update current commit parents
	if counter < len(places) {
		for i, p := range parents {
			res, err := updateCommit(p, copyMap(places), visited, counter, newMsg, repo)
			if err != nil {
				return nil, err
			}

			if parents[i].Id().String() != res.Id().String() {
				// update parent
				parents[i] = res
				changed = true
			}
		}
	}

	// commit has not been changed
	if !changed {
		visited[commit.Id().String()] = commit
		return commit, nil
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	oid, err := repo.CreateCommit(
		"",
		commit.Author(),
		commit.Committer(),
		message,
		tree,
		parents...
	)
	if err != nil {
		return nil, err
	}

	newCommit, err := GetCommit(oid.String(), repo)
	visited[commit.Id().String()] = newCommit
	return newCommit, err
}

func Update(
	repo *git.Repository,
	newMsg map[string]string,
) (*git.Commit, error) {
	// Prepare commits to be visited
	places := make(map[string]bool)
	for k, _ := range newMsg {
		places[k] = false
	}

	visited := make(map[string]*git.Commit)
	head, err := GetCommit("HEAD", repo)
	if err != nil {
		return nil, err
	}
	newHead, err := updateCommit(head, places, visited, 0, newMsg, repo)
	return newHead, err
}
