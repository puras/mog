package command

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func StopCmd() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop server",
		Action: func(c *cli.Context) error {
			fmt.Println("Stop Server")
			return nil
		},
	}
}
