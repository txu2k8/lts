package utils

import (
	"fmt"
	"os"

	"github.com/minio/cli"
)

// FlagToJSON converts a flag to a representation that can be reversed into the flag.
func FlagToJSON(ctx *cli.Context, flag cli.Flag) (string, error) {
	switch flag.(type) {
	case cli.StringFlag:
		if ctx.IsSet(flag.GetName()) {
			return ctx.String(flag.GetName()), nil
		}
	case cli.BoolFlag:
		if ctx.IsSet(flag.GetName()) {
			return fmt.Sprint(ctx.Bool(flag.GetName())), nil
		}
	case cli.Int64Flag:
		if ctx.IsSet(flag.GetName()) {
			return fmt.Sprint(ctx.Int64(flag.GetName())), nil
		}
	case cli.IntFlag:
		if ctx.IsSet(flag.GetName()) {
			return fmt.Sprint(ctx.Int(flag.GetName())), nil
		}
	case cli.DurationFlag:
		if ctx.IsSet(flag.GetName()) {
			return ctx.Duration(flag.GetName()).String(), nil
		}
	case cli.UintFlag:
		if ctx.IsSet(flag.GetName()) {
			return fmt.Sprint(ctx.Uint(flag.GetName())), nil
		}
	case cli.Uint64Flag:
		if ctx.IsSet(flag.GetName()) {
			return fmt.Sprint(ctx.Uint64(flag.GetName())), nil
		}
	case cli.Float64Flag:
		if ctx.IsSet(flag.GetName()) {
			return fmt.Sprint(ctx.Float64(flag.GetName())), nil
		}
	default:
		if ctx.IsSet(flag.GetName()) {
			return "", fmt.Errorf("unhandled flag type: %T", flag)
		}
	}
	return "", nil
}

// CommandLine attempts to reconstruct the commandline.
func CommandLine(ctx *cli.Context) string {
	s := os.Args[0] + " " + ctx.Command.Name
	for _, flag := range ctx.Command.Flags {
		val, err := FlagToJSON(ctx, flag)
		if err != nil || val == "" {
			continue
		}
		name := flag.GetName()
		switch name {
		case "access-key", "secret-key":
			val = "*REDACTED*"
		}
		s += " --" + flag.GetName() + "=" + val
	}
	return s
}
