package video

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path"
	"path/filepath"
	"stress/pkg/logger"

	"github.com/dustin/go-humanize"
)

func GetFileInfo(fullPath string) *FileInfo {
	fInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		logger.Errorf("File or Path not exist! -> %s", fullPath)
	}
	// 是否是目录
	if fInfo.IsDir() {
		files := GetDirFiles(fullPath)
		return GetFileInfo(files[0])
	}

	fSize := uint64(fInfo.Size())
	fName := fInfo.Name()
	fileInfo := FileInfo{
		Name:      fName,
		FullPath:  fullPath,
		FileType:  path.Ext(fName),
		Md5:       GetFileMd5(fullPath),
		Size:      fSize,
		SizeHuman: humanize.IBytes(fSize),
	}
	return &fileInfo
}

func GetDirFiles(rootDir string) []string {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if path != rootDir {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		logger.Debugf("%s", file)
	}
	return files
}

// 获取文件的md5码
func GetFileMd5(fullPath string) string {
	pFile, err := os.Open(fullPath)
	if err != nil {
		logger.Errorf("打开文件失败,path=%v, err=%v", fullPath, err)
		return ""
	}
	defer pFile.Close()
	md5h := md5.New()
	io.Copy(md5h, pFile)

	return hex.EncodeToString(md5h.Sum(nil))
}
