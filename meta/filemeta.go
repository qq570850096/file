package meta

import (
	"file_server/db"
	"sort"
	"sync"
)

// FileMeta: 文件元信息结构体
type FileMeta struct {
	FileSha1 string `json:"FileSha1"`
	FileName string `json:"FileName"`
	FileSize int64 `json:"FileSize"`
	Location string `json:"Location"`
	UploadAt string `json:"UploadAt"`
}

var fileMetas map[string]FileMeta
var once sync.Once
// 初始化时单例模式
func Init() map[string]FileMeta {
	once.Do (func() {
		fileMetas = make(map[string]FileMeta)
	})
	return fileMetas
}

// 新增或更新文件元信息
func UpdateFileMeta(fmeta FileMeta)  {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 获得文件元信息
func GetFileMeta(sha1 string) FileMeta {
	return fileMetas[sha1]
}

// 新增，更新文件元信息到mys ql中
func UpdateFileMetaDB(meta FileMeta) bool {
	return db.OnFileUploadFinished(meta.FileSha1,meta.FileName,meta.FileSize,meta.Location)
}

// GetFileMetaDB : 从mysql获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tfile, err := db.GetFileMeta(fileSha1)
	if tfile == nil || err != nil {
		return nil, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return &fmeta, nil
}

// GetLastFileMetas : 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}

	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// GetLastFileMetasDB : 批量从mysql获取文件元信息
func GetLastFileMetasDB(limit int) ([]FileMeta, error) {
	tfiles, err := db.GetFileMetaList(limit)
	if err != nil {
		return make([]FileMeta, 0), err
	}

	tfilesm := make([]FileMeta, len(tfiles))
	for i := 0; i < len(tfilesm); i++ {
		tfilesm[i] = FileMeta{
			FileSha1: tfiles[i].FileHash,
			FileName: tfiles[i].FileName.String,
			FileSize: tfiles[i].FileSize.Int64,
			Location: tfiles[i].FileAddr.String,
		}
	}
	sort.Sort(ByUploadTime(tfilesm))
	return tfilesm, nil
}

// RemoveFileMeta : 删除元信息
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}

