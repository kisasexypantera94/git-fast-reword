package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"

	"git-fast-reword/utilite"
)

func prettyMap(m map[string]string) (string, error) {
	pretty, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

func update(cfg map[string]string) error {
	newHashes, err := utilite.Update(".git", cfg)
	if err != nil {
		return err
	}
	pretty, err := prettyMap(newHashes)
	if err != nil {
		return err
	}
	fmt.Printf("New hashes:\n%s\n", pretty)
	return nil
}

func main() {
	app := &cli.App{
		Name:                 "git-fast-reword",
		Usage:                "fast commits rewording",
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

					data, err := ioutil.ReadFile(args.Get(0))
					if err != nil {
						return err
					}
					var cfg map[string]string
					err = json.Unmarshal(data, &cfg)
					if err != nil {
						return err
					}

					return update(cfg)
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

			return update(map[string]string{hash: msg})
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
}
