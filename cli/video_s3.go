package cli

import (
	s3client "stress/client/s3"
	"stress/pkg/workflow"
	"stress/pkg/workflow/video"

	"github.com/dustin/go-humanize"
	"github.com/minio/cli"
	"github.com/minio/minio-go/v7"
	"github.com/minio/pkg/console"
)

var videoBaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "channel-num",
		Value: 1,
		Usage: "videoBaseFlags: 业务模型 - 视频路数.",
	},
	cli.Float64Flag{
		Name:  "bitstream",
		Value: 4,
		Usage: "videoBaseFlags: 业务模型 - 视频码流大小(单位: Mbps).",
	},
	cli.Float64Flag{
		Name:  "datalife",
		Value: 0,
		Usage: "videoBaseFlags: 业务模型 - 视频保留期限(单位: 天), 0表示自动推算.",
	},
	cli.StringFlag{
		Name:  "capacity",
		Value: "0TiB",
		Usage: "videoBaseFlags: 业务模型 - 集群可用空间(数字或10KiB/MiB/GiB).",
	},
	cli.Float64Flag{
		Name:  "safe-water-level",
		Value: 0.91,
		Usage: "videoBaseFlags: 业务模型 - 集群安全水位.",
	},
	cli.StringFlag{
		Name:  "local-path",
		Value: "",
		Usage: "videoBaseFlags: 业务模型 - 指定源文件路径（指定目录）.",
	},
	cli.BoolFlag{
		Name:  "appendable",
		Usage: "videoBaseFlags: 业务模型 - 是否追加写模式.",
	},
	cli.IntFlag{
		Name:  "segments",
		Value: 1,
		Usage: "videoBaseFlags: 业务模型 - 追加写模式下，追加写入分片数(源文件分片后追加写入).",
	},
	cli.BoolFlag{
		Name:  "disable-multipart",
		Usage: "videoBaseFlags: 业务模型 - 是否非多段上传, 与appendable互斥, 即非追加写模式生效.",
	},
	cli.IntFlag{
		Name:  "max-workers",
		Value: 1,
		Usage: "videoBaseFlags: 业务模型 - 每路视频最大并发数.",
	},
	cli.IntFlag{
		Name:  "prepare-channel-num",
		Value: 1,
		Usage: "videoBaseFlags: 业务模型 - 数据预埋节点模拟视频路数（增大以加快预埋速度）.",
	},
}

var videoCustomFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "bucket-prefix",
		Value: "",
		Usage: "videoCustomFlags: 自定义 - 桶前缀.",
	},
	cli.StringFlag{
		Name:  "obj-prefix",
		Value: "",
		Usage: "videoCustomFlags: 自定义 - 对象前缀.",
	},
	cli.IntFlag{
		Name:  "idx-width",
		Value: 3,
		Usage: "videoCustomFlags: 自定义 - 对象序号长度,例如: 3=>001.",
	},
	cli.IntFlag{
		Name:  "idx-start",
		Value: 1,
		Usage: "videoCustomFlags: 自定义 - 对象序号起始值.",
	},
	cli.BoolFlag{
		Name:  "skip-stage-init",
		Usage: "videoCustomFlags: 自定义 - 跳过 init 阶段.",
	},
	cli.BoolFlag{
		Name:  "write-only",
		Usage: "videoCustomFlags: 自定义 -只写入，不删除.",
	},
	cli.BoolFlag{
		Name:  "delete-immediately",
		Usage: "videoCustomFlags: 自定义 -写入后，立即删除上一个.",
	},
	cli.BoolFlag{
		Name:  "single-root",
		Usage: "videoCustomFlags: 自定义 - 单个根目录/桶模式.",
	},
	cli.StringFlag{
		Name:  "single-root-name",
		Value: "video",
		Usage: "videoCustomFlags: 自定义 - 单根目录时，根目录名称.",
	},
	cli.IntFlag{
		Name:  "process-workers",
		Value: 8,
		Usage: "videoCustomFlags: 自定义 - 多进程运行协程，进程数.",
	},
	cli.IntFlag{
		Name:  "duration",
		Value: 0,
		Usage: "videoCustomFlags: 自定义 - 指定持续执行时间, 0-代表永久.",
	},
	cli.StringFlag{
		Name:   "part.size",
		Value:  "",
		Usage:  "videoCustomFlags: Multipart part size. Can be a number or 10KiB/MiB/GiB. All sizes are base 2 binary.",
		Hidden: true,
	},
}

// Video command.
var videoS3Cmd = cli.Command{
	Name:   "video-s3",
	Usage:  "video scene test: S3",
	Action: mainVideo,
	Before: setGlobalsFromContext,
	Flags:  combineFlags(aliasFlags, videoBaseFlags, videoCustomFlags, globalFlags),
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS]
  -> see https://github.com/txu2k8/storage-stress-test#video-s3

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}`,
}

// mainVideo is the entry point for cp command.
func mainVideo(ctx *cli.Context) error {
	checkVideoSyntax(ctx)
	capacity, _ := humanize.ParseBytes(ctx.String("capacity"))
	videoInfo := video.VideoInfo{
		VideoBaseInfo: video.VideoBaseInfo{
			FileInfo: video.FileInfo{
				FullPath: ctx.String("local-path"),
			},
			ChannelNum: ctx.Int("channel-num"),
			BitStream:  float32(ctx.Float64("bitstream")),
			DataLife:   float32(ctx.Float64("datalife")),
			// FileReader:       ,
			TotalCapacity:    capacity,
			SafeWaterLevel:   float32(ctx.Float64("safe-water-level")),
			Appendable:       ctx.Bool("appendable"),
			Segments:         ctx.Int("segments"),
			DisableMultipart: ctx.Bool("disable-multipart"),
			SingleBucket:     ctx.Bool("single-bucket"),
		},
	}
	videoInfo.CalcData()
	// src := newGenSource(ctx, "obj.size")
	b := video.VideoWorkflow{
		Common: workflow.Common{
			S3Client:    s3client.NewClient(ctx),
			Concurrency: ctx.Int("concurrent"),
		},
		VideoInfo: videoInfo,
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
