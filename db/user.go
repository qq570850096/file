package db

import (
	mydb "file_server/db/mysql"
	"fmt"
)

type User struct {
	Username string
	Email string
	Phone string
	SignupAt string
	LastActiveAt string
	Status int
}
// 通过用户名及密码的用户注册操作
func UserSignup(username, password string) bool {
	stmt,err := mydb.DBCONN().Prepare(
		"insert ignore into tbl_user(`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	defer stmt.Close()

	ret,err := stmt.Exec(username,password)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rowAffected,err := ret.RowsAffected();nil == err && rowAffected > 0 {
		return true
	}
	return false
}

// 登陆
func UserSignin(username,enc_pwd string) bool {
	stmt,err := mydb.DBCONN().Prepare(
		"select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows,err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}else if rows == nil {
		fmt.Println("username not found"+username)
		return false
	}
	// 返回相应数据字典
	pRow := mydb.ParseRows(rows)
	if len(pRow) > 0 && string(pRow[0]["user_pwd"].([]byte)) == enc_pwd {
		return true
	} else {
		return false
	}
}

func UpdateToken(username,token string) bool {
	stmt,err := mydb.DBCONN().Prepare(
		"replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	defer stmt.Close()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	_,err = stmt.Exec(username,token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 取用户信息
func GetUserInfo(username string)(User,error) {
	user := User{}

	stmt,err := mydb.DBCONN().Prepare(
		"select user_name,signup_at from tbl_user where user_name=? limit 1")
	defer stmt.Close()
	if err != nil {
		fmt.Println(err.Error())
		return User{}, err
	}
	// 执行查询
	err = stmt.QueryRow(username).Scan(&user.Username,&user.SignupAt)
	if err != nil {
		return user,err
	}
	return user,nil
}
