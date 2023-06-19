// 场景：视频监控 - 数据模型
package video

import "io"

// FileInfo 测试源文件信息
type FileInfo struct {
	Name      string        `json:"文件名"`       // 文件名
	FullPath  string        `json:"文件路径"`      // 文件路径
	FileType  string        `json:"文件类型"`      // 文件类型 -- 后缀
	Md5       string        `json:"MD5值"`      // MD5值
	Tags      string        `json:"Tags"`      // Tags
	Attr      string        `json:"Attr"`      // Attr
	Size      uint64        `json:"Size"`      // Size
	SizeHuman string        `json:"SizeHuman"` // SizeHuman
	Segments  int           `json:"分段数"`       // 分段数
	Reader    io.ReadSeeker `json:"Reader"`    // Reader
}

// VideoBaseInfo 原始需求信息
type VideoBaseInfo struct {
	ChannelNum int      `json:"视频路数"`       // 视频路数
	BitStream  float32  `json:"视频码流(Mbps)"` // 码流大小，单位 Mbps
	DataLife   float32  `json:"DataLife"`   // 数据保留期限，单位 天
	FileInfo   FileInfo `json:"FileInfo"`   // 源文件信息\Reader
	Segments   int      `json:"追加分片数"`      // 追加写模式下，一个对象追加分片次数

	TotalCapacity     uint64  `json:"TotalCapacity"`     // 存储池总容量大小，单位 byte
	SafeWaterLevel    float32 `json:"SafeWaterLevel"`    // 安全水位，即数据写入存储池的数据量最大不超过总容量的百分比，例如 90%=0.9
	SafeWaterCapacity uint64  `json:"SafeWaterCapacity"` // 安全水位存储池容量大小，单位 byte; SafeWaterCapacity=TotalCapacity*SafeWaterLevel

	Appendable       bool `json:"追加写模式"`  // 追加写模式
	DisableMultipart bool `json:"非多段上传"`  // 非多段上传，与追加写互斥
	SingleBucket     bool `json:"单桶模式"`   // 单桶模式，即所有数据存储在同一个桶中的不同路径
	SingleBucketName bool `json:"单桶模式桶名"` // 单桶模式，即所有数据存储在同一个桶中的不同路径

	// 仅用于打印
	FileInfoHuman          string `json:"源文件信息"`   // 源文件信息
	DataLifeHuman          string `json:"保留期限(天)"` // 数据保留期限，打印
	TotalCapacityHuman     string `json:"总容量"`     // 存储池总容量大小，打印
	SafeWaterLevelHuman    string `json:"安全水位"`    // 安全水位，打印
	SafeWaterCapacityHuman string `json:"安全容量"`    // 安全水位存储池容量大小，打印
}

// VideoDataInfo 数据模型信息 -- 原始需求信息分解后计算得出
type VideoDataInfo struct {
	BandWidth           float32 `json:"总带宽(MiB/s)"` // 带宽， MiB/s
	BucketNum           int     `json:"桶数量"`        // 桶数量
	ObjNum              int     `json:"总对象数"`       // 安全水位能写入的总对象数量，达到该数量后需要 边写边删
	ObjNumPC            int     `json:"每路视频对象数"`    // 安全水位每路视频能写入的对象数量
	ObjNumPD            int     `json:"每天对象数"`      // 每天需要写入的对象数量，码流+对象大小 计算得出
	ObjNumPCPD          int     `json:"每路视频每天对象数"`  // 每路视频、每天，需要写入的对象数量
	TimeInterval        int     `json:"一个对象产生时间"`   // 一路视频中，每个视频对象产生的实际间隔，单位：秒
	SegmentTimeInterval int     `json:"一个分片产生时间"`   // 一路视频中，每个视频对象分片产生的实际间隔，单位：秒

	MaxWorkers        int     `json:"一路视频最大并行数"` // 一路视频中，最大处理并行数
	MainConcurrent    float32 `json:"写删阶段每秒并行数"` // 写删阶段，平均每秒处理对象数
	PrepareConcurrent float32 `json:"预埋阶段每秒并行数"` // 数据预埋阶段，平均每秒处理对象数

}

// VideoCustomizeInfo 自定义变量信息
type VideoCustomizeInfo struct {
	PrepareChannelNum int    `json:"预埋阶段视频路数"` // 数据预埋阶段，指定视频路数模拟写入（为加快数据预埋），0=ChannelNum即和写删阶段相同
	BucketPrefix      string `json:"桶前缀"`      // 桶名 前缀
	ObjPrefix         string `json:"对象前缀"`     // 对象名 前缀
	ObjIdxWidth       int    `json:"对象名序号长度"`  // 对象名称序号长度， 3=>001
	ObjIdxStart       int    `json:"对象序号起始值"`  // 对象处理 序号起始值
}

// VideoInfo 视频监控场景 数据模型：原始需求信息 + 数据模型信息 + 自定义变量信息
type VideoInfo struct {
	VideoBaseInfo
	VideoDataInfo
	VideoCustomizeInfo
}
