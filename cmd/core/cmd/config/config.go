package config

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/wyt-labs/wyt-core/internal/pkg/config"
)

// Command is The explorer commands
var Command = &cli.Command{
	Name:        "config",
	Usage:       "The config manage commands",
	Subcommands: subCommands,
}

var subCommands = []*cli.Command{
	{
		Name:   "check",
		Usage:  "Check if the config file is valid",
		Action: check,
	},
	{
		Name:   "show",
		Usage:  "Show the complete config processed by the environment variable",
		Action: show,
	},
	{
		Name:   "generate-default",
		Usage:  "Generate default config",
		Action: generateDefault,
	},
}

func check(ctx *cli.Context) error {
	p, err := config.LoadRootPathFromEnv()
	if err != nil {
		fmt.Println("load root path env failed:", err)
		os.Exit(1)
		return nil
	}
	cfg := config.DefaultConfig(p)
	if err := config.ReadConfig(cfg); err != nil {
		fmt.Println("config file format error, please check:", err)
		os.Exit(1)
		return nil
	}
	return nil
}

func show(ctx *cli.Context) error {
	p, err := config.LoadRootPathFromEnv()
	if err != nil {
		fmt.Println("load root path env failed:", err)
		os.Exit(1)
		return nil
	}
	cfg := config.DefaultConfig(p)
	if err := config.ReadConfig(cfg); err != nil {
		fmt.Println("read config file failed:", err)
		os.Exit(1)
		return nil
	}
	str, err := config.MarshalConfig(cfg)
	if err != nil {
		fmt.Println("marshal config failed:", err)
		os.Exit(1)
		return nil
	}
	fmt.Println(str)
	return nil
}

func generateDefault(ctx *cli.Context) error {
	p, err := config.LoadRootPathFromEnv()
	if err != nil {
		fmt.Println("load root path env failed:", err)
		os.Exit(1)
		return nil
	}
	cfg := config.DefaultConfig(p)
	if config.IsConfigExist(cfg) {
		fmt.Println("config file already exists")
		os.Exit(1)
		return nil
	}

	if err := config.WriteConfig(cfg); err != nil {
		fmt.Println("write config to file failed:", err)
		os.Exit(1)
		return nil
	}
	return nil
}
