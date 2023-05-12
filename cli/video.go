package cli

import (
	"s3stress/client"
	"s3stress/pkg/bench"
	"s3stress/pkg/workflow/video"

	"github.com/minio/cli"
	"github.com/minio/minio-go/v7"
	"github.com/minio/pkg/console"
)

var videoFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "obj.size",
		Value: "10MiB",
		Usage: "putFlag: Size of each generated object. Can be a number or 10KiB/MiB/GiB. All sizes are base 2 binary.",
	},
	cli.StringFlag{
		Name:   "part.size",
		Value:  "",
		Usage:  "putFlag: Multipart part size. Can be a number or 10KiB/MiB/GiB. All sizes are base 2 binary.",
		Hidden: true,
	},
}

// Video command.
var videoCmd = cli.Command{
	Name:   "video",
	Usage:  "video scene test",
	Action: mainVideo,
	Before: setGlobalsFromContext,
	Flags:  combineFlags(aliasFlags, ioFlags, videoFlags, genFlags, benchFlags, analyzeFlags, globalFlags),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS]
  -> see https://github.com/txu2k8/s3stress#video

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}`,
}

// mainVideo is the entry point for cp command.
func mainVideo(ctx *cli.Context) error {
	checkVideoSyntax(ctx)
	src := newGenSource(ctx, "obj.size")
	b := video.VideoWorkflow{
		Common: bench.Common{
			Client:      client.NewClient(ctx),
			Concurrency: ctx.Int("concurrent"),
			Source:      src,
			Bucket:      ctx.String("bucket"),
			Location:    "",
			PutOpts:     videoOpts(ctx),
		},
	}
	return runBench(ctx, &b)
}

// videoOpts retrieves put options from the context.
func videoOpts(ctx *cli.Context) minio.PutObjectOptions {
	pSize, _ := toSize(ctx.String("part.size"))
	return minio.PutObjectOptions{
		ServerSideEncryption: newSSE(ctx),
		DisableMultipart:     ctx.Bool("disable-multipart"),
		SendContentMd5:       ctx.Bool("md5"),
		StorageClass:         ctx.String("storage-class"),
		PartSize:             pSize,
	}
}

func checkVideoSyntax(ctx *cli.Context) {
	if ctx.NArg() > 0 {
		console.Fatal("Command takes no arguments")
	}
}
