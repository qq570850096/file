package handler

import (
	"file_server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	rConn "file_server/cache/redis"
	"strings"
	"time"
	sql "file_server/db"
)

type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int  // 分块大小
	ChunkCount int  // 分块数量
}

// 初始化上传分块
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize,err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1,"paramas invalid", nil).JSONBytes())
	}
	// 2. 获得 redis 的一个连接
	conn := rConn.RedisPool().Get()
	defer conn.Close()

	// 3. 生成一个分块上传初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize)/(5 * 1024 * 1024))),
	}

	// 4.将初始化信息写入 Redis
	conn.Do("HSET","MP_"+upInfo.UploadID,"chunkcount",upInfo.ChunkCount)
	conn.Do("HSET","MP_"+upInfo.UploadID,"filehash",upInfo.FileHash)
	conn.Do("HSET","MP_"+upInfo.UploadID,"filesize",upInfo.FileSize)

	// 5. 将相应初始化数据返回客户端
	w.Write(util.NewRespMsg(0,"OK",upInfo).JSONBytes())
}

// CompleteUploadHandler : 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1.解析请求参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")
	//2.获得redis连接池中的一个连接
	conn := rConn.RedisPool().Get()
	defer conn.Close()
	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	data,err := conn.Do("HGETALLL","MP_"+upid)
	if err != nil {
		w.Write(util.NewRespMsg(-1,"compelete upload failed",nil).JSONBytes())
		return
	}
	datal := data.([]interface{})
	totalCount,chunkCount := 0,0
	for i:=0;i<len(datal);i+=2 {
		k := string(datal[i].([]byte))
		v := string(datal[i+1].([]byte))
		if k == "chunkcount" {
			totalCount,_ = strconv.Atoi(v)
		}else if strings.HasPrefix(k,"chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-1,"invalid request",nil).JSONBytes())
		return
	}
	// 4.TODO:合并分块

	// 5.更新唯一文件表及用户文件表
	fsize,_ := strconv.Atoi(filesize)
	sql.OnUserFileUploadFinished(username,filehash,filename,int64(fsize))
	sql.OnFileUploadFinished(filehash,filename,int64(fsize),"")
	// 6.进行回应
	w.Write(util.NewRespMsg(0,"OK",nil).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	// username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadID")
	chunkIDX := r.Form.Get("index")

	// 2. 获得一个 redis 连接
	conn := rConn.RedisPool().Get()
	defer conn.Close()
	// 3. 获得文件句柄，用于存储分块内容
	fpath := "/data/"+uploadID+"/"+chunkIDX
	os.MkdirAll(path.Dir(fpath),0744)
	file,err := os.Create(fpath)
	defer file.Close()
	if err!=nil {
		w.Write(util.NewRespMsg(-1,"Upload Part Failed",nil).JSONBytes())
		return
	}
	buf := make([]byte,1024*1024)
	for {
		n,err := r.Body.Read(buf)
		file.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 4. 更新redis缓存状态
	conn.Do("HSET","MP_"+uploadID,"chkidx_"+chunkIDX,1)

	// 5. 返回处理结果给客户端
	w.Write(util.NewRespMsg(0,"OK",nil).JSONBytes())
}