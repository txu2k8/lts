// 场景：视频监控 - 数据模型
package video

// VideoBaseInfo 原始需求信息
type VideoBaseInfo struct {
	ChannelNum int     // 视频路数
	BitStream  int     // 码流大小，单位 Mbps
	DataLife   float32 // 数据保留期限，单位 天
	ObjSize    int     // 对象大小，单位 MiB

	TotalCapacity  int     // 存储池总容量大小，单位 MiB
	SafeWaterLevel float32 // 安全水位，即数据写入存储池的数据量最大不超过总容量的百分比，例如 90%=0.9

	Appendable   bool // 追加写模式
	Segments     int  // 追加写模式下，一个对象追加分片次数
	Multipart    bool // 多段上传，与追加写互斥
	SingleBucket bool // 单桶模式，即所有数据存储在同一个桶中的不同路径
}

// VideoDataInfo 数据模型信息 -- 原始需求信息分解后计算得出
type VideoDataInfo struct {
	VideoBaseInfo
	SafeWaterCapacity int // 存储池安全水位容量大小，单位 MiB

	BandWidth    int // 带宽， MiB/s
	BucketNum    int // 桶数量
	ObjNum       int // 安全水位能写入的总对象数量，达到该数量后需要 边写边删
	ObjNumPC     int // 安全水位每路视频能写入的对象数量
	ObjNumPD     int // 每天需要写入的对象数量，码流+对象大小 计算得出
	ObjNumPCPD   int // 每路视频、每天，需要写入的对象数量
	TimeInterval int // 一路视频中，没个视频对象产生的实际间隔，单位：秒
	MaxWorkers   int // 一路视频中，最大处理并行数

	PrepareConcurrent float32 // 数据预埋阶段，平均每秒处理对象数
	MainConcurrent    float32 // 写删阶段，平均每秒处理对象数
}

// VideoCustomizeInfo 自定义变量信息
type VideoCustomizeInfo struct {
	PrepareChannelNum int    // 数据预埋阶段，指定视频路数模拟写入（为加快数据预埋），0=ChannelNum即和写删阶段相同
	BucketPrefix      string // 桶名 前缀
	ObjPrefix         int    // 对象名 前缀
	ObjIdxWidth       int    // 对象名称序号长度， 3=>001
	ObjIdxStart       int    // 对象处理 序号起始值
}

// VideoInfo 视频监控场景 数据模型：原始需求信息 + 数据模型信息 + 自定义变量信息
type VideoInfo struct {
	VideoBaseInfo
	VideoDataInfo
	VideoCustomizeInfo
}
