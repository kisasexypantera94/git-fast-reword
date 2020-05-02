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
	counter int,
	children map[string][]string,
	newMsg map[string]string,
	repo *git.Repository,
) (*git.Commit, error) {
	// Check if current commit is to be renamed
	cid := commit.Id().String()
	if visited, ok := places[cid]; ok && !visited {
		counter += 1
	}

	parents := getParents(commit)
	// if there are still commits to be visited then do recursion
	// and update current commit parents
	if counter < len(places) {
		for i, p := range parents {
			pid := p.Id().String()
			children[pid] = append(children[pid], commit.Id().String())
			res, err := updateCommit(p, copyMap(places), counter, children, newMsg, repo)
			if err != nil {
				return nil, err
			}
			// update parent
			parents[i] = res
		}
	}

	// Get message and update if needed
	message := commit.Message()
	if newMsg, ok := newMsg[commit.Id().String()]; ok {
		message = newMsg
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
	return newCommit, err
}

func Update(
	hash string,
	repo *git.Repository,
	newMsg map[string]string,
) (*git.Commit, error) {
	children := make(map[string][]string)
	places := map[string]bool{hash: false}
	head, err := GetCommit("HEAD", repo)
	if err != nil {
		return nil, err
	}
	newHead, err := updateCommit(head, places, 0, children, newMsg, repo)
	return newHead, err
}
