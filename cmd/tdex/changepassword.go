package main

import (
	"context"
	"fmt"

	daemonv2 "github.com/tdex-network/tdex-daemon/api-spec/protobuf/gen/tdex-daemon/v2"
	"github.com/urfave/cli/v2"
)

const (
	curPwdFlagName = "current_password"
	newPwdFlagName = "new_password"
)

var changepassword = cli.Command{
	Name:  "changepassword",
	Usage: "change the password to unlock the wallet of the daemon. Requires the daemon to be locked.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  curPwdFlagName,
			Usage: "the old unlocking password to be changed",
			Value: "",
		},
		&cli.StringFlag{
			Name:  "new_password",
			Usage: "the new password that replaces the old one",
			Value: "",
		},
	},
	Action: changePasswordAction,
}

func changePasswordAction(ctx *cli.Context) error {
	client, cleanup, err := getWalletClient(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	curPwd := ctx.String(curPwdFlagName)
	newPwd := ctx.String(newPwdFlagName)

	if _, err := client.ChangePassword(context.Background(), &daemonv2.ChangePasswordRequest{
		CurrentPassword: curPwd,
		NewPassword:     newPwd,
	}); err != nil {
		return err
	}

	fmt.Println("Done")

	return nil
}
