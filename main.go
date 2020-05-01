package main

import (
	"fmt"
	"os"

	"github.com/libgit2/git2go/v30"
)

func getParents(commit *git.Commit) []*git.Commit {
	parents := make([]*git.Commit, commit.ParentCount())
	for i := 0; i < len(parents); i++ {
		parents[i] = commit.Parent(uint(i))
	}
	return parents
}

func mapParentsToChildren(start string, repo *git.Repository) (map[string]string, error) {
	walker, err := repo.Walk()
	if err != nil {
		return nil, err
	}

	// Map parents to children
	children := make(map[string]string)
	err = walker.PushRange(fmt.Sprintf("%s..HEAD", start))
	if err != nil {
		return nil, err
	}
	err = walker.Iterate(func(commit *git.Commit) bool {
		parents := getParents(commit)
		for _, p := range parents {
			pid := p.Id().String()
			children[pid] = commit.Id().String()
		}
		return true
	})

	return children, err
}

func getCommit(hash string, repo *git.Repository) (*git.Commit, error) {
	obj, err := repo.RevparseSingle(hash)
	if err != nil {
		return nil, err
	}
	commit, err := obj.AsCommit()
	return commit, err
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
	//fmt.Println(hash, oid.String())
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
	hash := os.Args[1]
	msg := os.Args[2]

	repo, err := git.OpenRepository(".git")
	if err != nil {
		panic(err)
	}

	obj, _ := getCommit(hash, repo)
	c, _ := obj.AsCommit()

	start := c.Id().String()
	newMsg := make(map[string]string)
	newMsg[start] = msg

	children, err := mapParentsToChildren(start, repo)
	if err != nil {
		panic(err)
	}
	newHead, err := update(start, "", "", repo, children, newMsg)
	if err != nil {
		panic(err)
	}

	fmt.Println(newHead)

	ref, err := repo.References.Lookup("refs/heads/master")
	if err != nil {
		panic(err)
	}
	_, err = ref.SetTarget(newHead, "")
	if err != nil {
		panic(err)
	}
}
