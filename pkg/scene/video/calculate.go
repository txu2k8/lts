package video_scense

// 打印计算结果
func (v *VideoInfo) printVideoInfo() bool {
	// TODO
	return true
}

// 视频监控场景 - 数据模型计算
func (v *VideoInfo) CalcData() *VideoInfo {
	v.SafeWaterCapacity = int(float32(v.TotalCapacity) * v.SafeWaterLevel)
	if v.PrepareChannelNum < v.ChannelNum {
		v.PrepareChannelNum = v.ChannelNum
	}

	// 总带宽=码流/8*路数
	v.BandWidth = (v.BitStream / 8) * v.ChannelNum

	// 对象数 = 安全水位容量 / 对象大小
	v.ObjNum = v.SafeWaterCapacity / v.ObjSize

	// 每天数据量 = 带宽 * 1天
	var sizePD = v.BandWidth * 60 * 60 * 24
	if v.DataLife == 0 {
		v.DataLife = float32(v.SafeWaterCapacity / sizePD)
	}

	// 每天一路视频需要写入的数据量
	var sizePCPD = (v.BitStream / 8) * 60 * 60 * 24

	// 更加码流+容量+保留期限，换算 支持的视频路数
	if v.ChannelNum == 0 && v.DataLife > 0 {
		v.ChannelNum = int(float32(v.SafeWaterCapacity)/v.DataLife) / sizePCPD
	}

	// 打印计算结果
	v.printVideoInfo()
	return v
}
