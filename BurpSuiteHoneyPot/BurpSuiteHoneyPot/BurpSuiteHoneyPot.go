package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"github.com/julienschmidt/httprouter"
)

//判断文件或目录是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}



//开启http服务
func main() {
	if fileExists("./html") {
		port := flag.String("p", "80", "Input http port")
		flag.Parse()

		filename:=filepath.Base(os.Args[0])
		fmt.Printf("use %s -port [port] to run \n default port is 80 \n",filename)

		p, h := NewFileHandle("/")
		r := httprouter.New()
		r.GET(p, LogHandle(h))

		log.Fatalln(http.ListenAndServe(":"+*port, r))
	}else{
		fmt.Println("请检查html文件夹是否存在")
	}
}


func NewFileHandle(path string) (string, httprouter.Handle) {
	return fmt.Sprintf("%s*files", path), func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		http.StripPrefix(path, http.FileServer(http.Dir("./html"))).ServeHTTP(w, r)
	}
}


func LogHandle(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		now := time.Now()
		handle(w, r, p)
		file := "./" + time.Now().Format("2006-01-02") + ".txt"
		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		// 设置日志输出到文件
		// 定义多个写入器
		writers := []io.Writer{
			f,
			os.Stdout}
		fileAndStdoutWriter := io.MultiWriter(writers...)
		// 创建新的log对象
		logger := log.New(fileAndStdoutWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
		// 使用新的log对象，写入日志内容
		logger.Printf("%s %s %s done in %v", r.RemoteAddr, r.Method, r.URL.Path, time.Since(now))
		logger.Println("Request User-Agent:", r.Header.Get("User-Agent"))
	}
}