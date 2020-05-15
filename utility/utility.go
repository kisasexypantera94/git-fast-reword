package utility

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

func convertToHashes(repo *git.Repository, m map[string]string) (map[string]string, map[string]string, error) {
	newMsg := make(map[string]string)
	translator := make(map[string]string)
	for k, v := range m {
		c, err := getCommit(k, repo)
		if err != nil {
			return nil, nil, err
		}
		newMsg[c.Id().String()] = v + "\n"
		translator[c.Id().String()] = k
	}
	return newMsg, translator, nil
}

// Map new commit hashes back to commit names from config
func convertFromHashes(newHashes map[string]string, translator map[string]string) map[string]string {
	translatedBack := make(map[string]string)
	for k, v := range translator {
		translatedBack[v] = newHashes[k]
	}
	return translatedBack
}

func getCommit(hash string, repo *git.Repository) (*git.Commit, error) {
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
	commit *git.Commit,             // current commit
	places map[string]bool,         // commits to be renamed
	visited map[string]*git.Commit, // visited commits cache
	counter int,                    // current number of places visited
	newMsg map[string]string,       // new messages for commits
	newHashes map[string]string,    // new hashes of reworded commits
	repo *git.Repository,
) (*git.Commit, error) {
	// Check if current commit was already updated
	if newCommit, ok := visited[commit.Id().String()]; ok {
		return newCommit, nil
	}

	// Check if current commit is to be renamed
	cid := commit.Id().String()
	if renamed, ok := places[cid]; ok && !renamed {
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

	// if there are still commits to be renamed then do recursion
	// and update current commit parents
	parents := getParents(commit)
	if counter < len(places) {
		for i := range parents {
			// TODO: iterative
			res, err := updateCommit(
				parents[i],
				copyMap(places),
				visited,
				counter,
				newMsg,
				newHashes,
				repo,
			)
			if err != nil {
				return nil, err
			}

			// update parent
			if parents[i].Id().String() != res.Id().String() {
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
	// create new commit with updated meta
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

	newCommit, err := getCommit(oid.String(), repo)
	// cache new commit
	visited[commit.Id().String()] = newCommit
	newHashes[commit.Id().String()] = oid.String()

	return newCommit, err
}

func Update(
	repoPath string,
	cfg map[string]string,
) (map[string]string, error) {
	// Open repository
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}
	// Convert commits to hashes if needed
	newMsg, translator, err := convertToHashes(repo, cfg)
	if err != nil {
		return nil, err
	}
	// Prepare commits to be visited
	places := make(map[string]bool)
	for k := range newMsg {
		places[k] = false
	}
	// Get current HEAD
	head, err := getCommit("HEAD", repo)
	if err != nil {
		return nil, err
	}
	visited := make(map[string]*git.Commit)
	newHashes := make(map[string]string)
	newHead, err := updateCommit(
		head,
		places,
		visited,
		0,
		newMsg,
		newHashes,
		repo,
	)
	if err != nil {
		return nil, err
	}

	// Get HEAD reference
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}
	// Update it
	_, err = ref.SetTarget(newHead.Id(), "")

	return convertFromHashes(newHashes, translator), err
}
