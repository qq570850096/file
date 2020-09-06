package handler

import (
	db "file_server/db"
	"file_server/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// 用于加密
const (
	pwd_salt = "*&%890"
)
// 处理用户注册请求的请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data,err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Write(data)
			return
		}
	}
	r.ParseForm()
	
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	enc_passwd := util.Sha1([]byte(password+pwd_salt))
	if db.UserSignup(username,enc_passwd) {
		resp := util.RespMsg{
			Code: 200,
			Msg:  "OK",
			Data: "SUCCESS",
		}
		w.Write(resp.JSONBytes())
	} else {
		w.Write([]byte("FAILED"))
	}

}
// 登陆接口
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signin.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	enc_passwd := util.Sha1([]byte(password+pwd_salt))
	// TODO：校验用户名密码
	pwdChecked := db.UserSignin(username,enc_passwd)
	// 如果检查不通过
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}
	// TODO: 生成访问凭证
	token := GenToken(username)
	if !db.UpdateToken(username,token) {
		resp := util.RespMsg{
			Code: 400,
			Msg:  "NO",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// TODO: 登陆成功后重定向到首页
	resp := util.RespMsg{
		Code: 200,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token string
		}{
			Location:"../main/main",
			Username:username,
			Token:token,
		},
	}
	w.Write(resp.JSONBytes())
}

// 生成Token
func GenToken(username string) string {
	// 40位的token
	ts := fmt.Sprintf("%x",time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username+ts+"_tokensalt"))
	return tokenPrefix + ts[:8]
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: 解析得到请求参数
	r.ParseForm()
	username:=r.Form.Get("username")
	token := r.Form.Get("token")
	// TODO：判断TOKEN是否有效
	flag := IsTokenValid(token)
	if !flag {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 查询用户信息
	user,err := db.GetUserInfo(username)
	if err!= nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 组装并相应用户数据
	resp := util.RespMsg{
		Code: 200,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

func IsTokenValid(token string) bool {
	// TODO: 判断TOken是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息

	// TODO: 对比两个token是否一致

	return true
}

func HomeHandler(w http.ResponseWriter,r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/home.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
}
