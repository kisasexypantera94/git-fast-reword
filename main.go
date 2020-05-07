package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/libgit2/git2go/v30"
	"github.com/urfave/cli/v2"

	"git-fast-reword/utilite"
)

func updateAndSetRef(repo *git.Repository, newMsg map[string]string) error {
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
	return err
}

func main() {
	app := &cli.App{
		Name:                 "git-fast-reword",
		Usage:                "git-fast-reword hash new_message",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "from-file",
				Aliases: []string{"ff"},
				Action: func(c *cli.Context) error {
					args := c.Args()
					if args.Len() < 1 {
						return fmt.Errorf("invalid number of arguments")
					}

					repo, err := git.OpenRepository(".git")
					if err != nil {
						return err
					}

					data, err := ioutil.ReadFile(args.Get(0))
					if err != nil {
						return err
					}
					var cfg map[string]string
					err = json.Unmarshal(data, &cfg)
					if err != nil {
						return err
					}

					newMsg := make(map[string]string)
					for k, v := range cfg {
						c, err := utilite.GetCommit(k, repo)
						if err != nil {
							return err
						}
						newMsg[c.Id().String()] = v + "\n"
					}

					return updateAndSetRef(repo, newMsg)
				},
			},
		},
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

			newMsg := make(map[string]string)
			newMsg[commit.Id().String()] = msg + "\n"

			return updateAndSetRef(repo, newMsg)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}
