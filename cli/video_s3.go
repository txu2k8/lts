package cli

import (
	s3client "stress/client/s3"
	"stress/models"
	"stress/pkg/logger"
	"stress/workflow"
	"stress/workflow/video"

	"github.com/dustin/go-humanize"
	"github.com/minio/cli"
	"github.com/minio/minio-go/v7"
	"github.com/minio/pkg/console"
)

var videoBaseFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "channel-num",
		Value: 1,
		Usage: "业务模型 - 视频路数.",
	},
	cli.Float64Flag{
		Name:  "bitstream",
		Value: 4,
		Usage: "业务模型 - 视频码流大小(单位: Mbps).",
	},
	cli.Float64Flag{
		Name:  "datalife",
		Value: 0,
		Usage: "业务模型 - 视频保留期限(单位: 天), 0表示自动推算.",
	},
	cli.StringFlag{
		Name:  "capacity",
		Value: "0TiB",
		Usage: "业务模型 - 集群可用空间(数字或10KiB/MiB/GiB).",
	},
	cli.Float64Flag{
		Name:  "safe-water-level",
		Value: 0.91,
		Usage: "业务模型 - 集群安全水位.",
	},
	cli.StringFlag{
		Name:  "local-path",
		Value: "",
		Usage: "业务模型 - 指定源文件路径（指定目录）.",
	},
	cli.BoolFlag{
		Name:  "appendable",
		Usage: "业务模型 - 是否追加写模式.",
	},
	cli.IntFlag{
		Name:  "appendable.segments",
		Value: 1,
		Usage: "业务模型 - 追加写模式下，追加写入分片数(源文件分片后追加写入).",
	},
	cli.BoolFlag{
		Name:  "disable-multipart",
		Usage: "业务模型 - 非多段上传, 与appendable互斥, 即非追加写模式生效.",
	},
	cli.StringFlag{
		Name:  "multipart.part_size",
		Value: "",
		Usage: "业务模型 - 多段上传时, part size. Can be a number or 10KiB/MiB/GiB. All sizes are base 2 binary.",
		// Hidden: true,
	},
	cli.BoolFlag{
		Name:  "md5",
		Usage: "业务模型 - Add MD5 sum to uploads",
		// Hidden: true,
	},
	cli.IntFlag{
		Name:  "max-workers",
		Value: 1,
		Usage: "业务模型 - 每路视频最大并发数.",
	},
	cli.IntFlag{
		Name:  "prepare-channel-num",
		Value: 1,
		Usage: "业务模型 - 数据预埋节点模拟视频路数（增大以加快预埋速度）.",
	},
}

var videoCustomFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "bucket-prefix",
		Value: "",
		Usage: "自定义 - 桶前缀.",
	},
	cli.StringFlag{
		Name:  "obj-prefix",
		Value: "",
		Usage: "自定义 - 对象前缀.",
	},
	cli.IntFlag{
		Name:  "idx-width",
		Value: 3,
		Usage: "自定义 - 对象序号长度,例如: 3=>001.",
	},
	cli.IntFlag{
		Name:  "idx-start",
		Value: 1,
		Usage: "自定义 - 对象序号起始值.",
	},
	cli.BoolFlag{
		Name:  "skip-stage-init",
		Usage: "自定义 - 跳过创建桶阶段.",
	},
	cli.BoolFlag{
		Name:  "write-only",
		Usage: "自定义 - 只写入，不删除.",
	},
	cli.BoolFlag{
		Name:  "delete-immediately",
		Usage: "自定义 - 写入后，立即删除上一个.",
	},
	cli.BoolFlag{
		Name:  "single-root",
		Usage: "自定义 - 单个根目录/桶模式.",
	},
	cli.StringFlag{
		Name:  "single-root.name",
		Value: "video",
		Usage: "自定义 - 单根目录时，根目录名称.",
	},
	cli.IntFlag{
		Name:  "process-workers",
		Value: 8,
		Usage: "自定义 - 多进程运行协程，进程数.",
	},
	cli.IntFlag{
		Name:  "duration",
		Value: 0,
		Usage: "自定义 - 指定持续执行时间, 0-代表永久.",
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
	// 初始化zap logger
	logger.InitLogger("video_s3", "text", "debug", true)
	checkVideoSyntax(ctx)
	capacity, _ := humanize.ParseBytes(ctx.String("capacity"))
	videoInfo := video.VideoInfo{
		VideoBaseInfo: video.VideoBaseInfo{
			FileInfo: models.FileInfo{
				FullPath: ctx.String("local-path"),
			},
			ChannelNum: ctx.Int("channel-num"),
			BitStream:  float32(ctx.Float64("bitstream")),
			DataLife:   float32(ctx.Float64("datalife")),
			// FileReader:       ,
			TotalCapacity:    capacity,
			SafeWaterLevel:   float32(ctx.Float64("safe-water-level")),
			Appendable:       ctx.Bool("appendable"),
			Segments:         ctx.Int("appendable.segments"),
			DisableMultipart: ctx.Bool("disable-multipart"),
			SingleBucket:     ctx.Bool("single-root"),
			SingleBucketName: ctx.Bool("single-root.name"),
		},
	}
	videoInfo.CalcData()
	// src := newGenSource(ctx, "obj.size")
	b := video.VideoWorkflow{
		Common: workflow.Common{
			S3Client:    s3client.NewClient(ctx),
			Concurrency: ctx.Int("concurrent"),
			PutOpts:     videoPutOpts(ctx),
		},
		VideoInfo: videoInfo,
	}
	return workflow.RunWorkflow(ctx, &b)
}

// videoPutOpts retrieves put options from the context.
func videoPutOpts(ctx *cli.Context) minio.PutObjectOptions {
	pSize, _ := humanize.ParseBytes(ctx.String("multipart.part_size"))
	return minio.PutObjectOptions{
		ServerSideEncryption: newSSE(ctx),
		DisableMultipart:     ctx.Bool("disable-multipart"),
		SendContentMd5:       ctx.Bool("md5"),
		PartSize:             pSize,
	}
}

func checkVideoSyntax(ctx *cli.Context) {
	if ctx.NArg() > 0 {
		console.Fatal("Command takes no arguments")
	}
}
