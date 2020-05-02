package main

import (
	"fmt"
	"os"

	"github.com/libgit2/git2go/v30"
	"github.com/urfave/cli/v2"
)

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

// iterateUntilVisited fills children map.
// It does DFS until  all the places are visited
// and then returns oldest place visited.
func iterateUntilVisited(
	commit *git.Commit,
	places map[string]bool,
	counter int,
	children map[string]string,
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
		children[pid] = commit.Id().String()
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

func mapParentsToChildren(start string, repo *git.Repository) (map[string]string, *git.Commit, error) {
	children := make(map[string]string)
	places := map[string]bool{start: false}
	head, err := getCommit("HEAD", repo)
	if err != nil {
		return nil, nil, err
	}
	oldest, err := iterateUntilVisited(head, places, 0, children)
	return children, oldest, err
}

func update(
	hash string,
	parent string,
	oldParent string,
	repo *git.Repository,
	children map[string]string,
	newMsg map[string]string,
) (*git.Oid, error) {
	commit, err := getCommit(hash, repo)
	if err != nil {
		return nil, err
	}

	message := commit.Message()
	if newMsg, ok := newMsg[hash]; ok {
		message = newMsg
	}

	parents := getParents(commit)
	for i, p := range parents {
		if p.Id().String() == oldParent {
			pCommit, err := getCommit(parent, repo)
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
	oid, err := repo.CreateCommit("", commit.Author(), commit.Committer(), message, tree, parents...)
	if err != nil {
		return nil, err
	}

	newHead := oid
	if ch, ok := children[hash]; ok {
		res, err := update(ch, oid.String(), hash, repo, children, newMsg)
		if err != nil {
			return nil, err
		}
		if res != nil {
			newHead = res
		}
	}

	return newHead, nil
}

func main() {
	app := &cli.App{
		Name:                 "git-fast-reword",
		Usage:                "git-fast-reword hash new_message",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			args := c.Args()
			if args.Len() < 2 {
				return fmt.Errorf("invalid number of arguments")
			}
			hash := args.Get(0)
			msg := args.Get(1)

			repo, err := git.OpenRepository(".git")
			if err != nil {
				return err
			}

			obj, err := getCommit(hash, repo)
			if err != nil {
				return err
			}
			commit, err := obj.AsCommit()
			if err != nil {
				return err
			}

			start := commit.Id().String()
			newMsg := make(map[string]string)
			newMsg[start] = msg

			children, oldest, err := mapParentsToChildren(start, repo)
			if err != nil {
				return err
			}
			newHead, err := update(oldest.Id().String(), "", "", repo, children, newMsg)
			if err != nil {
				return err
			}

			fmt.Println(newHead)

			ref, err := repo.References.Lookup("refs/heads/master")
			if err != nil {
				return err
			}
			_, err = ref.SetTarget(newHead, "")
			if err != nil {
				return err
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
