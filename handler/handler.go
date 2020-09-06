package handler

import (
	"encoding/json"
	db "file_server/db"
	"file_server/meta"
	"file_server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 定义上传接口
func UploadHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}
		io.WriteString(w, string(data))
		// 另一种返回方式:
		// 动态文件使用http.HandleFunc设置，静态文件使用到http.FileServer设置(见main.go)
		// 所以直接redirect到http.FileServer所配置的url
		// http.Redirect(w, r, "/static/view/index.html",  http.StatusFound)
	} else if r.Method == "POST" {
		// 接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data, err:%s\n", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileSha1: "",
			FileName: head.Filename,
			FileSize: 0,
			Location: "./tmp/"+head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}
		meta.Init()

		// 按照文件名进行存储
		newfile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Println("Failed to create file",err)
		}
		defer  newfile.Close()

		fileMeta.FileSize, err = io.Copy(newfile,file)
		if err != nil {
			fmt.Println("Failed to save data into file",err)
		}
		// 将文件句柄定位到起始位置
		newfile.Seek(0,0)
		fileMeta.FileSha1 = util.FileSha1(newfile)
		// meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// TODO: 更新用戶文件表記錄
		r.ParseForm()
		username := r.Form.Get("username")
		suc := db.OnUserFileUploadFinished(username,fileMeta.FileSha1,fileMeta.FileName,fileMeta.FileSize)
		if suc {
			// 重定向到文件上传成功
			http.Redirect(w,r,"/file/upload/suc",http.StatusFound)
		} else {
			w.Write([]byte("UpLoad Failed!"))
		}


	}
}

func UploadSucHandler(w http.ResponseWriter, r *http.Request)  {
	http.Redirect(w,r,"/home",http.StatusFound)
}

// GetFileMetaHandler : 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta,err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 注意私有属性是不能被序列化的
	data,err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	// fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// DownloadHandler : 文件下载接口
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm,err := meta.GetFileMetaDB(fsha1)
	if err!=nil {
		fmt.Println(err.Error())
		w.Write([]byte("DownLoad Failed"))
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	//data, err := ioutil.ReadAll(f)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//data = data

	w.Header().Add("Content-Type", "application/octect-stream")
	//// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Add("content-disposition", "attachment; filename=\""+fm.FileName+"\"")
	http.ServeFile(w, r, fm.Location)
}

// FileMetaUpdateHandler ： 更新元信息接口(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	// TODO: 更新文件表中的元信息记录

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler : 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")

	fMeta := meta.GetFileMeta(fileSha1)
	// 删除文件
	os.Remove(fMeta.Location)
	// 删除文件元信息
	meta.RemoveFileMeta(fileSha1)
	// TODO: 删除表文件信息

	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler:尝试秒传
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize,_ := strconv.Atoi(r.Form.Get("filesize"))

	meta,err := meta.GetFileMetaDB(filehash)
	if err!=nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if meta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，之前并没有这个文件",
			Data: nil,
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		suc := db.OnUserFileUploadFinished(username,filehash,filename,int64(filesize))
		if suc {
			resp := util.RespMsg{
				Code: 0,
				Msg:  "秒传成功",
				Data: nil,
			}
			w.Write(resp.JSONBytes())
			return
		} else {
			resp := util.RespMsg{
				Code: -2,
				Msg:  "秒传失败,服务正忙，请稍后重试",
				Data: nil,
			}
			w.Write(resp.JSONBytes())
			return
		}
	}
}
