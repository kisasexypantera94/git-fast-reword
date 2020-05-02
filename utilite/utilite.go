package utilite

import git "github.com/libgit2/git2go/v30"

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

// iterateUntilVisited fills children map.
// It does DFS until  all the places are visited
// and then returns oldest place visited.
func iterateUntilVisited(
	commit *git.Commit,
	places map[string]bool,
	counter int,
	children map[string][]string,
) (oldest *git.Commit, err error) {
	cid := commit.Id().String()
	if visited, ok := places[cid]; ok && !visited {
		counter += 1
	}

	if counter == len(places) {
		return commit, nil
	}

	parents := getParents(commit)
	for _, p := range parents {
		pid := p.Id().String()
		children[pid] = append(children[pid], commit.Id().String())
		pc, err := p.AsCommit()
		if err != nil {
			return nil, err
		}
		oldest, err = iterateUntilVisited(pc, places, counter, children)
		if err != nil {
			return nil, err
		}
		if oldest != nil {
			return oldest, nil
		}
	}

	return nil, nil
}

func MapParentsToChildren(start string, repo *git.Repository) (map[string][]string, *git.Commit, error) {
	children := make(map[string][]string)
	places := map[string]bool{start: false}
	head, err := GetCommit("HEAD", repo)
	if err != nil {
		return nil, nil, err
	}
	oldest, err := iterateUntilVisited(head, places, 0, children)
	return children, oldest, err
}

func Update(
	hash string,
	parent string,
	oldParent string,
	repo *git.Repository,
	children map[string][]string,
	newMsg map[string]string,
) (*git.Oid, error) {
	// Get current commit
	commit, err := GetCommit(hash, repo)
	if err != nil {
		return nil, err
	}

	// Get message and update if needed
	message := commit.Message()
	if newMsg, ok := newMsg[hash]; ok {
		message = newMsg
	}

	// Get parents and update old parent hash
	parents := getParents(commit)
	for i, p := range parents {
		if p.Id().String() == oldParent {
			pCommit, err := GetCommit(parent, repo)
			if err != nil {
				return nil, err
			}
			parents[i] = pCommit
		}
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

	// Mark created commit as current head
	newHead := oid
	for _, ch := range children[hash] {
		res, err := Update(ch, oid.String(), hash, repo, children, newMsg)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newHead = res
		}
	}

	return newHead, nil
}
