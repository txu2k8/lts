package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/cli"
	"github.com/posener/complete"
)

// Main function to answer to bash completion calls
func mainComplete() error {
	// Recursively register all commands and subcommands
	// along with global and local flags
	complCmds := make(complete.Commands)
	for _, cmd := range appCmds {
		complCmds[cmd.Name] = cmdToCompleteCmd(cmd, "")
	}
	complFlags := flagsToCompleteFlags(nil)
	cliComplete := complete.Command{
		Sub:         complCmds,
		GlobalFlags: complFlags,
	}
	// Answer to bash completion call
	complete.New(filepath.Base(os.Args[0]), cliComplete).Run()
	return nil
}

// The list of all commands supported by warp with their mapping
// with their bash completer function
var completeCmds = map[string]complete.Predictor{
	"/version": nil,
}

// flagsToCompleteFlags transforms a cli.Flag to complete.Flags
// understood by posener/complete library.
func flagsToCompleteFlags(flags []cli.Flag) complete.Flags {
	complFlags := make(complete.Flags)
	for _, f := range flags {
		for _, s := range strings.Split(f.GetName(), ",") {
			var flagName string
			s = strings.TrimSpace(s)
			if len(s) == 1 {
				flagName = "-" + s
			} else {
				flagName = "--" + s
			}
			complFlags[flagName] = complete.PredictNothing
		}
	}
	return complFlags
}

// This function recursively transforms cli.Command to complete.Command
// understood by posener/complete library.
func cmdToCompleteCmd(cmd cli.Command, parentPath string) complete.Command {
	var complCmd complete.Command
	complCmd.Sub = make(complete.Commands)

	for _, subCmd := range cmd.Subcommands {
		complCmd.Sub[subCmd.Name] = cmdToCompleteCmd(subCmd, parentPath+"/"+cmd.Name)
	}

	complCmd.Flags = flagsToCompleteFlags(cmd.Flags)
	complCmd.Args = completeCmds[parentPath+"/"+cmd.Name]
	return complCmd
}
