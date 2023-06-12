package video

import (
	"context"
	"fmt"
	"net/http"
	"stress/pkg/bench"
	"stress/pkg/utils"
	"stress/pkg/workflow"
	"sync"
	"time"
)

const (
	// 定义一个开始日期字符串
	dateString = "2023-1-1"
	// 定义日期的格式
	layout = "2006-01-02"
)

// Put benchmarks upload speed.
type VideoWorkflow struct {
	workflow.Common
	VideoInfo
	prefixes          map[string]struct{}
	ChannelID         int    // 视频ID
	ChannelName       string // 视频Name
	SkipStageInit     bool   //跳过init阶段
	WriteOnly         bool   //只写
	DeleteImmediately bool   // 立即删除
	SingleRoot        bool   // 单桶模式
	SingleRootName    string // 单桶名称
	Duration          int    // 指定运行时间

	Depth int // 目录深度，默认1

}

// calc_date_string 计算日期下一天
func (u *VideoWorkflow) calc_date_string(startDate string, dateStep int) string {
	// 解析字符串为 time.Date 类型
	t, _ := time.Parse(layout, startDate)
	// 获取指定日期之后N天的日期
	t = t.AddDate(0, 0, dateStep)
	return t.Format(layout)
}

func (u *VideoWorkflow) calc_obj_prefix(objPrefix string, depth int, datePrefix string) string {
	if datePrefix == "today" {
		datePrefix = time.Now().Format(layout) + "/"
	}
	nestedPrefix := ""
	for d := 1; d < depth; d++ {
		nestedPrefix += fmt.Sprintf("nested%d/", d)
	}
	return datePrefix + nestedPrefix + objPrefix + fmt.Sprintf("-ch%d", u.ChannelID)
}

// Calc_obj_path 计算对象path
func (u *VideoWorkflow) Calc_obj_path(idx int) string {
	dateStep := idx / u.ObjNumPCPD
	datePrefix := u.calc_date_string(dateString, dateStep) + "/"
	filePrefix := u.calc_obj_prefix(u.ObjPrefix, u.Depth, datePrefix)
	filePath := filePrefix + utils.Zfill(string(idx), u.ObjIdxWidth) // + file_type
	if u.SingleRoot {
		filePath = fmt.Sprintf("%s/%s", u.ChannelName, filePath)
	}

	return filePath
}

// Producer
func (u *VideoWorkflow) Producer(ch chan int, count int, wg *sync.WaitGroup) {
	defer close(ch)
	wg.Add(count)
	for i := 0; i < count; i++ {
		ch <- i
		fmt.Println("任务", i, "生产完毕")
	}
}

// Consumer
func (u *VideoWorkflow) Consumer(ch chan int, pool chan struct{}, wg *sync.WaitGroup) {
	for c := range ch {
		pool <- struct{}{}
		<-pool
		go u.Worker(c, wg)
	}
}

// 具体消费逻辑
func (u *VideoWorkflow) Worker(c int, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("任务", c, "消费完毕")
}

func (u *VideoWorkflow) StageMain() {
	//初始化管道来接收任务数据
	ch := make(chan int, 10000)
	//所有任务执行完毕才结束进程
	wg := &sync.WaitGroup{}
	//用来控制协程数量,超过50个会阻塞
	pool := make(chan struct{}, 50)
	//任务数量
	count := 1000
	go u.Producer(ch, count, wg)
	u.Consumer(ch, pool, wg)
	wg.Wait()
	fmt.Println("任务处理完毕")
}

// Prepare will create an empty buckets ot delete any content already there.
func (u *VideoWorkflow) Prepare(ctx context.Context) error {
	// 输入参数计算分析
	u.CalcData()

	fmt.Printf("Stage-Init:Create empty buckets: %s%d~%d", u.BucketPrefix, 0, u.BucketNum)
	return nil // u.CreateEmptyBucket(ctx)
}

// Start will execute the main workflow.
// Operations should begin executing when the start channel is closed.
func (u *VideoWorkflow) Start(ctx context.Context, wait chan struct{}) (bench.Operations, error) {
	var wg sync.WaitGroup
	wg.Add(u.Concurrency)
	c := bench.NewCollector()
	if u.AutoTermDur > 0 {
		ctx = c.AutoTerm(ctx, http.MethodPut, u.AutoTermScale, bench.AutoTermCheck, bench.AutoTermSamples, u.AutoTermDur)
	}
	u.prefixes = make(map[string]struct{}, u.Concurrency)

	// Non-terminating context.
	nonTerm := context.Background()

	for i := 0; i < u.Concurrency; i++ {
		src := u.Source()
		u.prefixes[src.Prefix()] = struct{}{}
		go func(i int) {
			rcv := c.Receiver()
			defer wg.Done()
			opts := u.PutOpts
			done := ctx.Done()

			<-wait
			for {
				select {
				case <-done:
					return
				default:
				}
				obj := src.Object()
				opts.ContentType = obj.ContentType
				client, cldone := u.S3Client()
				op := bench.Operation{
					OpType:   http.MethodPut,
					Thread:   uint16(i),
					Size:     obj.Size,
					File:     obj.Name,
					ObjPerOp: 1,
					Endpoint: client.EndpointURL().String(),
				}
				op.Start = time.Now()
				res, err := client.PutObject(nonTerm, u.Bucket, obj.Name, obj.Reader, obj.Size, opts)
				op.End = time.Now()
				if err != nil {
					u.Error("upload error: ", err)
					op.Err = err.Error()
				}
				obj.VersionID = res.VersionID

				if res.Size != obj.Size && op.Err == "" {
					err := fmt.Sprint("short upload. want:", obj.Size, ", got:", res.Size)
					if op.Err == "" {
						op.Err = err
					}
					u.Error(err)
				}
				op.Size = res.Size
				cldone()
				rcv <- op
			}
		}(i)
	}
	wg.Wait()
	return c.Close(), nil
}

// Cleanup deletes everything uploaded to the bucket.
func (u *VideoWorkflow) Cleanup(ctx context.Context) {
	var pf []string
	for p := range u.prefixes {
		pf = append(pf, p)
	}
	u.DeleteAllInBucket(ctx, pf...)
}
