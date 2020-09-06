package db

import (
	mysql "file_server/db/mysql"
	"fmt"
	"time"
)

// 用戶文件表結構體
type UserFile struct {
	Username string
	FileHash string
	FileName string
	FileSize int64
	UploadAt string
	LastUpdated string
}

func OnUserFileUploadFinished(username, filehash, filename string, filesize int64) bool {
	stmt,err:=mysql.DBCONN().Prepare(
		"insert ignore into tbl_user_file (`user_name`,`file_sha1`,`file_name`," +
			"`file_size`,`upload_at`) values(?,?,?,?,?)")
	defer stmt.Close()
	if err!=nil {
		return false
	}
	// exec返回的是影響的行數，Query 返回的是對象列表
	_,err = stmt.Exec(username,filehash,filename,filesize,time.Now())
	if err != nil {
		return false
	}
	return true
}

// 查詢用戶文件關聯
func QueryUserFileMetas(username string, limit int) ([]UserFile,error) {
	stmt,err := mysql.DBCONN().Prepare(
		"select file_sha1, file_name, file_size, upload_at,last_update from tbl_user_file where user_name=? limit ?")
	defer stmt.Close()
	if err != nil {
		return nil,err
	}

	rows,err := stmt.Query(username,limit)
	if err!=nil {
		return nil,err
	}
	userFiles := make([]UserFile,0)
	for rows.Next() {
		ufile := UserFile{}
		err = rows.Scan(&ufile.FileHash,&ufile.FileName,&ufile.FileSize,
			&ufile.UploadAt,&ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		userFiles = append(userFiles,ufile)
	}
	return userFiles,nil
}