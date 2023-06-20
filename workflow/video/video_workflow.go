package video

import (
	"fmt"
	"stress/pkg/utils"
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
	VideoInfo
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
	filePath := filePrefix + utils.Zfill(fmt.Sprint(idx), u.ObjIdxWidth) // + file_type
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
