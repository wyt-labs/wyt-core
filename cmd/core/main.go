package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/wyt-labs/wyt-core/cmd/core/cmd"
	configcmd "github.com/wyt-labs/wyt-core/cmd/core/cmd/config"
	"github.com/wyt-labs/wyt-core/internal/pkg/config"
)

func main() {
	app := cli.NewApp()
	app.Name = config.AppName
	app.Usage = "wyt is a web3 project analysis platform"
	app.HideVersion = true
	app.Description = "Run COMMAND --help for more information on a command"

	app.Commands = []*cli.Command{
		{
			Name:   "start",
			Usage:  "Start app",
			Action: cmd.Start,
		},
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "Show version",
			Action: func(c *cli.Context) error {
				config.PrintSystemInfo("", func(format string, args ...any) {
					fmt.Printf(format+"\n", args...)
				})
				return nil
			},
		},
		configcmd.Command,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("app run error: %v\n", err)
	}
}
