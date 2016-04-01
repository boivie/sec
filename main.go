package main

import (
	"os"
	"github.com/codegangsta/cli"
	"github.com/boivie/sec/cmd"
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: "db",
			Value: "testdb",
			Usage: "Path to database",
		},
	}
}

func getCommands() []cli.Command {
	return []cli.Command{
		cmd.CmdInit,
		cmd.CmdDump,
		cmd.CmdServe,
		cmd.CmdAuditor,
		cmd.CmdOfferIdentity,
		cmd.CmdClaimIdentity,
		cmd.CmdIssueIdentity,
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "sec"
	app.Usage = "Secure identification"
	app.Flags = getFlags()
	app.Commands = getCommands()
	app.Run(os.Args)
}
