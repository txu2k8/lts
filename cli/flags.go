package cli

import (
	"fmt"
	"s3stress/client"
	"s3stress/config"

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
	config.GlobalQuiet = config.GlobalQuiet || quiet
	config.GlobalDebug = config.GlobalDebug || debug
	config.GlobalJSON = config.GlobalJSON || json
	config.GlobalNoColor = config.GlobalNoColor || noColor

	// Disable colorified messages if requested.
	if config.GlobalNoColor || config.GlobalQuiet {
		console.SetColorOff()
	}
}

// Flags common across all I/O commands such as cp, mirror, stat, pipe etc.
var aliasFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "endpoint",
		Usage:  "aliasFlag: endpoint. Multiple endpoints can be specified as a comma separated list.",
		EnvVar: config.AppNameUC + "_ENDPOINT",
		Value:  "127.0.0.1:6600",
	},
	cli.StringFlag{
		Name:   "access-key",
		Usage:  "aliasFlag: Specify access key",
		EnvVar: config.AppNameUC + "_ACCESS_KEY",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "secret-key",
		Usage:  "aliasFlag: Specify secret key",
		EnvVar: config.AppNameUC + "_SECRET_KEY",
		Value:  "",
	},
	cli.BoolFlag{
		Name:   "tls",
		Usage:  "aliasFlag: Use TLS (HTTPS) for transport",
		EnvVar: config.AppNameUC + "_TLS",
	},
	cli.StringFlag{
		Name:   "region",
		Usage:  "aliasFlag: Specify a custom region",
		EnvVar: config.AppNameUC + "_REGION",
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
		Value: string(client.HostSelectTypeWeighed),
		Usage: fmt.Sprintf("aliasFlag: Host selection algorithm. Can be %q or %q", client.HostSelectTypeWeighed, client.HostSelectTypeRoundrobin),
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
		Value: config.AppName + "-benchmark-bucket",
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
