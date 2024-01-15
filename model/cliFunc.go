package model

import "github.com/urfave/cli/v2"

type CliFunc func(cCtx *cli.Context) error
