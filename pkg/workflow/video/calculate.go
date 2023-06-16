package video

import (
	"fmt"
	"reflect"
	"strconv"
	"stress/pkg/logger"
	"strings"

	"github.com/dustin/go-humanize"
)

// ForeachStruct 遍历并打印结构体
func foreachStruct(obj interface{}) {
	max := 30
	t := reflect.TypeOf(obj) // 注意，obj不能为指针类型，否则会报：panic recovered: reflect: NumField of non-struct type
	v := reflect.ValueOf(obj)
	for k := 0; k < t.NumField(); k++ {
		tag := t.Field(k).Tag.Get("json")
		value := v.Field(k).Interface()
		switch tag {
		case "FileInfo", "DataLife", "TotalCapacity", "SafeWaterCapacity", "SafeWaterLevel":
			continue
		default:
			logger.Infof("%-"+strconv.Itoa(max)+"s\t: %v", tag, value)
		}
	}
}

// 打印计算结果
func (v *VideoInfo) printVideoInfo() bool {
	fmtStr := strings.Repeat("=", 30)
	logger.Infof("%s 原始需求信息 %s", fmtStr, fmtStr)
	foreachStruct(v.VideoBaseInfo)

	logger.Infof("%s 数据模型信息 %s", fmtStr, fmtStr)
	foreachStruct(v.VideoDataInfo)

	logger.Infof("%s 自定义变量信息 %s", fmtStr, fmtStr)
	foreachStruct(v.VideoCustomizeInfo)

	return true
}

// 视频监控场景 - 数据模型计算
func (v *VideoInfo) CalcData() *VideoInfo {
	logger.Infof("计算分析数据模型/参数...")
	v.FileInfo = *GetFileInfo(v.FileInfo.FullPath)
	v.FileInfoHuman = fmt.Sprintf("Path=%s; Size=%s", v.FileInfo.FullPath, v.FileInfo.SizeHuman)
	v.TotalCapacityHuman = humanize.IBytes(v.TotalCapacity)
	v.SafeWaterLevelHuman = fmt.Sprintf("%v %%", v.SafeWaterLevel*100)
	v.SafeWaterCapacity = uint64(float32(v.TotalCapacity) * v.SafeWaterLevel)
	v.SafeWaterCapacityHuman = humanize.IBytes(v.SafeWaterCapacity)
	if v.PrepareChannelNum < v.ChannelNum {
		v.PrepareChannelNum = v.ChannelNum
	}

	// 总带宽=码流/8*路数
	v.BandWidth = (v.BitStream / 8) * float32(v.ChannelNum)

	// 对象数 = 安全水位容量 / 对象大小
	v.ObjNum = int(v.SafeWaterCapacity / v.FileInfo.Size)

	// 每天数据量 = 带宽 * 1天
	var sizePD = v.BandWidth * 60 * 60 * 24 * 1024 * 1024
	if v.DataLife == 0 {
		// v.DataLife = float32(utils.Decimal(float64(v.SafeWaterCapacity/1024/1024/uint64(sizePD)), -2))
		v.DataLife = float32(v.SafeWaterCapacity) / sizePD
	}
	v.DataLifeHuman = fmt.Sprintf("%.3f", v.DataLife)

	// 每天一路视频需要写入的数据量
	var sizePCPD = sizePD / float32(v.ChannelNum)

	// 更加码流+容量+保留期限，换算 支持的视频路数
	if v.ChannelNum == 0 && v.DataLife > 0 {
		v.ChannelNum = int((float32(v.SafeWaterCapacity) / v.DataLife) / sizePCPD)
	}

	// 打印计算结果
	v.printVideoInfo()
	return v
}
