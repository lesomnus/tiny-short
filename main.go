package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/lesomnus/tiny-short/cmd"
	"github.com/lesomnus/tiny-short/cmd/key"
	"github.com/urfave/cli/v2"
)

func main() {
	var conf *cmd.Config
	color.NoColor = false

	app := &cli.App{
		Name: "tiny-short",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "conf",
				Aliases: []string{"c"},
				Value:   ".tiny-short.yaml",
				Usage:   "path to a config file",
			},
		},
		Before: func(c *cli.Context) error {
			p := c.String("conf")
			conf_, err := cmd.ReadConfig(p)
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			conf = conf_
			return nil
		},
		Action: func(c *cli.Context) error {
			if conf == nil {
				panic("config not loaded")
			}

			return cmd.Root(context.Background(), conf)
		},

		Commands: []*cli.Command{
			{
				Name:  "key",
				Usage: "utilities for keys",

				Subcommands: []*cli.Command{
					key.Gen,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
