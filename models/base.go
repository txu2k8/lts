package models

// 基本公用数据模型

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
