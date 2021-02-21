package main

import (
	"github.com/urfave/cli/v2"
)

func commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:   "yml",
			Usage:  "📄 Render helmwave.yml.tpl -> helmwave.yml",
			Action: app.CliYml,
			Before:  app.InitApp,
		},
		planCommand(),
		{
			Name:    "deploy",
			Aliases: []string{"apply", "sync", "release"},
			Usage:   "🛥 Deploy your helmwave!",
			Action:  app.CliDeploy,
			Before:  app.InitApp,
		},
		{
			Name:    "manifest",
			Aliases: []string{"manifest"},
			Usage:   "🛥 Fake Deploy",
			Action:  app.CliManifests,
			Before:  app.InitApp,
		},
		{
			Name: "version",
			Usage: "Print helmwave version",
			Action: app.CliVersion,
		},
	}

}

func planCommand() *cli.Command {
	return &cli.Command{
		Name: "planfile",
		Aliases: []string{"plan"},
		Usage:   "📜 Generate planfile to plandir",
		Before:  app.InitApp,
		Subcommands: []*cli.Command{
			{
				Name: "repo",
				Action: app.CliPlan,
			},
			{
				Name: "releases",
				Action: app.CliPlan,
			},
			{
				Name: "values",
				Action: app.CliPlan,
			},
			{
				Name: "all",
				Action: app.CliPlan,
			},
		},
	}
}

func help(c *cli.Context) error {
	args := c.Args()
	if args.Present() {
		return cli.ShowCommandHelp(c, args.First())
	}

	return cli.ShowAppHelp(c)
}
