package main

import (
	"fmt"
	"github.com/libgit2/git2go/v30"
)

func getParents(commit *git.Commit) []*git.Commit {
	parents := make([]*git.Commit, commit.ParentCount())
	for i := 0; i < len(parents); i++ {
		parents[i] = commit.Parent(uint(i))
	}
	return parents
}

func mapParentsToChildren(repo *git.Repository) (map[string][]string, error) {
	walker, err := repo.Walk()
	if err != nil {
		return nil, err
	}

	err = walker.PushRef("HEAD")
	if err != nil {
		return nil, err
	}

	// Map parents to children
	children := make(map[string][]string)
	err = walker.Iterate(func(commit *git.Commit) bool {
		parents := getParents(commit)
		for _, p := range parents {
			pid := p.Id().String()
			children[pid] = append(children[pid], commit.Id().String())
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

func foo(
	hash string,
	parent string,
	oldParent string,
	repo *git.Repository,
	children map[string][]string,
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
	for _, ch := range children[hash] {
		res, err := foo(ch, oid.String(), hash, repo, children, newMsg)
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
	//argsWithoutProg := os.Args[1:]

	repo, err := git.OpenRepository("/Users/chingachgook/dev/django-like-queryset/.git")
	//repo, err := git.OpenRepository("intellij-community/.git")
	if err != nil {
		panic(err)
	}

	children, err := mapParentsToChildren(repo)
	if err != nil {
		panic(err)
	}

	fmt.Println(children)

	start := "d313b6038578ecd90ef41aa8c8bc64fcb6889662"
	newMsg := make(map[string]string)
	newMsg["d313b6038578ecd90ef41aa8c8bc64fcb6889662"] = "iluv git"
	newHead, err := foo(start, "", "", repo, children, newMsg)
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

	//fmt.Println(children)
}
