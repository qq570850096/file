package handler

import "net/http"

// HTTPInterceptor : http请求拦截器
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			token := r.Form.Get("token")

			//验证登录token是否有效
			if  !IsTokenValid(token) {
				// w.WriteHeader(http.StatusForbidden)
				// token校验失败则跳转到登录页面
				http.Redirect(w, r, "/user/signin", http.StatusFound)
				return
			}
			h(w, r)
		})
}
