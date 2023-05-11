package cli

import (
	"fmt"
	"os"

	"github.com/minio/cli"
	"github.com/minio/pkg/console"
)

// Collection of warp flags currently supported
var globalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "quiet, q",
		Usage:  "disable progress bar display",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:  "no-color",
		Usage: "disable color theme",
	},
	cli.BoolFlag{
		Name:   "json",
		Usage:  "enable JSON formatted output",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:  "debug",
		Usage: "enable debug output",
	},
	cli.BoolFlag{
		Name:  "insecure",
		Usage: "disable TLS certificate verification",
	},
	cli.BoolFlag{
		Name:  "autocompletion",
		Usage: "install auto-completion for your shell",
	},
}

var profileFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "pprofdir",
		Usage:  "Write profiles to this folder",
		Value:  "pprof",
		Hidden: true,
	},

	cli.BoolFlag{
		Name:   "cpu",
		Usage:  "Write a local CPU profile",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:   "mem",
		Usage:  "Write an local allocation profile",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:   "block",
		Usage:  "Write a local goroutine blocking profile",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:   "mutex",
		Usage:  "Write a mutex contention profile",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:   "threads",
		Usage:  "Write a threas create profile",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:   "trace",
		Usage:  "Write an local execution trace",
		Hidden: true,
	},
}

// Set global states. NOTE: It is deliberately kept monolithic to ensure we dont miss out any flags.
func setGlobalsFromContext(ctx *cli.Context) error {
	quiet := ctx.IsSet("quiet")
	debug := ctx.IsSet("debug")
	json := ctx.IsSet("json")
	noColor := ctx.IsSet("no-color")
	setGlobals(quiet, debug, json, noColor)
	return nil
}

// Set global states. NOTE: It is deliberately kept monolithic to ensure we dont miss out any flags.
func setGlobals(quiet, debug, json, noColor bool) {
	globalQuiet = globalQuiet || quiet
	globalDebug = globalDebug || debug
	globalJSON = globalJSON || json
	globalNoColor = globalNoColor || noColor

	// Disable colorified messages if requested.
	if globalNoColor || globalQuiet {
		console.SetColorOff()
	}
}

// commandLine attempts to reconstruct the commandline.
func commandLine(ctx *cli.Context) string {
	s := os.Args[0] + " " + ctx.Command.Name
	for _, flag := range ctx.Command.Flags {
		val, err := flagToJSON(ctx, flag)
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

// Flags common across all I/O commands such as cp, mirror, stat, pipe etc.
var aliasFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "endpoint",
		Usage:  "aliasFlag: endpoint. Multiple endpoints can be specified as a comma separated list.",
		EnvVar: appNameUC + "_ENDPOINT",
		Value:  "127.0.0.1:6600",
	},
	cli.StringFlag{
		Name:   "access-key",
		Usage:  "aliasFlag: Specify access key",
		EnvVar: appNameUC + "_ACCESS_KEY",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "secret-key",
		Usage:  "aliasFlag: Specify secret key",
		EnvVar: appNameUC + "_SECRET_KEY",
		Value:  "",
	},
	cli.BoolFlag{
		Name:   "tls",
		Usage:  "aliasFlag: Use TLS (HTTPS) for transport",
		EnvVar: appNameUC + "_TLS",
	},
	cli.StringFlag{
		Name:   "region",
		Usage:  "aliasFlag: Specify a custom region",
		EnvVar: appNameUC + "_REGION",
		Hidden: true,
	},
	cli.StringFlag{
		Name:   "signature",
		Usage:  "aliasFlag: Specify a signature method. Available values are S3V2, S3V4",
		Value:  "S3V4",
		Hidden: true,
	},
	cli.StringFlag{
		Name:  "host-select",
		Value: string(hostSelectTypeWeighed),
		Usage: fmt.Sprintf("aliasFlag: Host selection algorithm. Can be %q or %q", hostSelectTypeWeighed, hostSelectTypeRoundrobin),
	},
	cli.BoolFlag{
		Name:   "resolve-host",
		Usage:  "aliasFlag: Resolve the host(s) ip(s) (including multiple A/AAAA records). This can break SSL certificates, use --insecure if so",
		Hidden: true,
	},
}

// Flags common across all I/O commands such as cp, mirror, stat, pipe etc.
var ioFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "encrypt",
		Usage: "ioFlags: encrypt/decrypt objects (using server-side encryption with random keys)",
	},
	cli.StringFlag{
		Name:  "bucket",
		Value: appName + "-benchmark-bucket",
		Usage: "ioFlags: Bucket to use for benchmark data. ALL DATA WILL BE DELETED IN BUCKET!",
	},
	cli.IntFlag{
		Name:  "concurrent",
		Value: 20,
		Usage: "ioFlags: Run this many concurrent operations",
	},
	cli.BoolFlag{
		Name:  "noprefix",
		Usage: "ioFlags: Do not use separate prefix for each thread",
	},
	cli.StringFlag{
		Name:  "prefix",
		Usage: "ioFlags: Use a custom prefix for each thread",
	},
	cli.BoolFlag{
		Name:  "disable-multipart",
		Usage: "ioFlags: disable multipart uploads",
	},
	cli.BoolFlag{
		Name:  "md5",
		Usage: "ioFlags: Add MD5 sum to uploads",
	},
	cli.StringFlag{
		Name:  "storage-class",
		Value: "",
		Usage: "ioFlags: Specify custom storage class, for instance 'STANDARD' or 'REDUCED_REDUNDANCY'.",
	},
	cli.BoolFlag{
		Name:   "disable-http-keepalive",
		Usage:  "ioFlags: Disable HTTP Keep-Alive",
		Hidden: true,
	},
}
