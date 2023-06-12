package cli

import (
	"stress/client"
	"stress/pkg/workflow"
	"stress/pkg/workflow/video"

	"github.com/minio/cli"
	"github.com/minio/minio-go/v7"
	"github.com/minio/pkg/console"
)

var videoFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "channel-num",
		Value: 1,
		Usage: "videoFlag: 业务模型 - 视频路数.",
	},
	cli.IntFlag{
		Name:  "bitstream",
		Value: 4,
		Usage: "videoFlag: 业务模型 - 视频码流大小(单位: Mbps).",
	},
	cli.Float64Flag{
		Name:  "datalife",
		Value: 4,
		Usage: "videoFlag: 业务模型 - 视频保留期限(单位: 天), 0表示自动推算.",
	},
	cli.StringFlag{
		Name:  "obj-size",
		Value: "128MiB",
		Usage: "videoFlag: 业务模型 - 对象大小(数字或10KiB/MiB/GiB). ",
	},
	cli.StringFlag{
		Name:  "capacity",
		Value: "1TiB",
		Usage: "videoFlag: 业务模型 - 集群可用空间(数字或10KiB/MiB/GiB).",
	},
	cli.Float64Flag{
		Name:  "safe-level",
		Value: 0.91,
		Usage: "videoFlag: 业务模型 - 集群安全水位.",
	},
	cli.BoolFlag{
		Name:  "appendable",
		Usage: "videoFlag: 业务模型 - 是否追加写模式.",
	},
	cli.IntFlag{
		Name:  "segments",
		Value: 1,
		Usage: "videoFlag: 业务模型 - 追加写模式下，追加写入分片数(最终对象大小=obj-size*segments).",
	},
	cli.BoolFlag{
		Name:  "multipart",
		Usage: "videoFlag: 业务模型 - 是否多段上传, 与appendable互斥, 即非追加写模式生效.",
	},
	cli.StringFlag{
		Name:   "part.size",
		Value:  "",
		Usage:  "videoFlag: Multipart part size. Can be a number or 10KiB/MiB/GiB. All sizes are base 2 binary.",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:  "single-bucket",
		Usage: "videoFlag: 业务模型 - 单个桶模式.",
	},
}

// Video command.
var videoCmd = cli.Command{
	Name:   "video",
	Usage:  "video scene test",
	Action: mainVideo,
	Before: setGlobalsFromContext,
	Flags:  combineFlags(aliasFlags, videoFlags, globalFlags),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS]
  -> see https://github.com/txu2k8/stress#video

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}`,
}

// mainVideo is the entry point for cp command.
func mainVideo(ctx *cli.Context) error {
	checkVideoSyntax(ctx)
	// src := newGenSource(ctx, "obj.size")
	b := video.VideoWorkflow{
		Common: workflow.Common{
			Client:      client.NewClient(ctx),
			Concurrency: ctx.Int("concurrent"),
		},
		VideoInfo: video.VideoInfo{
			VideoBaseInfo: video.VideoBaseInfo{
				ChannelNum:     ctx.Int("channel-num"),
				BitStream:      ctx.Int("bitstream"),
				DataLife:       float32(ctx.Float64("datalife")),
				ObjSize:        ctx.Int("obj.size"),
				TotalCapacity:  ctx.Int("capacity"),
				SafeWaterLevel: float32(ctx.Float64("safe-level")),
				Appendable:     ctx.Bool("appendable"),
				Segments:       ctx.Int("segments"),
				Multipart:      ctx.Bool("multipart"),
				SingleBucket:   ctx.Bool("single-bucket"),
			},
		},
	}
	return workflow.RunWorkflow(ctx, &b)
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
