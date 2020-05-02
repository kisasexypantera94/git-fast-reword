package main

import (
	"fmt"
	"log"
	"os"

	"github.com/libgit2/git2go/v30"
	"github.com/urfave/cli/v2"

	"git-fast-reword/utilite"
)

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

			commit, err := utilite.GetCommit(hash, repo)
			if err != nil {
				return err
			}

			start := commit.Id().String()
			newMsg := make(map[string]string)
			newMsg[start] = msg + "\n"

			newHead, err := utilite.Update(repo, newMsg)
			if err != nil {
				return err
			}

			log.Printf("New head: %s", newHead.Id().String())

			ref, err := repo.Head()
			if err != nil {
				return err
			}
			_, err = ref.SetTarget(newHead.Id(), "")
			if err != nil {
				return err
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}
