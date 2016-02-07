package cmd
import "github.com/codegangsta/cli"

var CmdAuditor = cli.Command{
	Name:      "auditor",
	Usage:     "options for task templates",
	Action: func(c *cli.Context) {
		println("completed task: ", c.Args().First())
	},
}
