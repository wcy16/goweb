package main

import (
	"./router"
	"fmt"
	"net/http"
)

func getroot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "get, /")
}

func gettest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "get, /test")
}

func getid(w http.ResponseWriter, r *http.Request) {
	param, ok := r.Context().Value("param").(map[string]string)
	if !ok {
		fmt.Fprintf(w, "wrong in getid")
	} else {
		fmt.Fprintf(w, "get, /"+param["id"])
	}
}

func getval(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "get, /other/val")
}

func main() {

	route := router.NewRoute()
	route.POST("/#id", getid)
	route.GET("/", getroot)
	route.GET("/test", gettest)
	route.GET("/#id", getid)
	route.GET("/#id/val", getid)
	route.GET("/other/val", getval)

	http.ListenAndServe(":9999", route)
}
