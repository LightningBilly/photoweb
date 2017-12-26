package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	ListDir      = 0x0001
	UPLOAD_DIR   = "./uploads"
	TEMPLATE_DIR = "./views"
)

var templates = make(map[string]*template.Template)

//初始化页面模板
func init() {
	fileInfoArr, err := ioutil.ReadDir(TEMPLATE_DIR)
	if err != nil {
		panic(err)
		return
	}

	var templateName, templatePath string
	for _, fileInfo := range fileInfoArr {
		templateName = fileInfo.Name()
		if ext := path.Ext(templateName); ext != ".html" {
			continue
		}

		templatePath = TEMPLATE_DIR + "/" + templateName
		log.Println("Loading template:", templatePath)
		t := template.Must(template.ParseFiles(templatePath))
		tmpl := getFileName(templateName)
		templates[tmpl] = t
	}
}

//
func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flags int) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		file := staticDir + r.URL.Path[len(prefix)-1:]
		if (flags & ListDir) == 0 {
			if exists := isExists(file); !exists {
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, file)
	})
}

//获取文件名（去除后缀）
func getFileName(fullFileName string) string {
	//fullFilename := "test.txt"
	//fmt.Println("fullFilename =", fullFilename)
	var filenameWithSuffix string
	filenameWithSuffix = path.Base(fullFileName)
	//fmt.Println("filenameWithSuffix =", filenameWithSuffix)
	var fileSuffix string
	fileSuffix = path.Ext(filenameWithSuffix)
	//fmt.Println("fileSuffix =", fileSuffix)

	var fileNameOnly string
	fileNameOnly = strings.TrimSuffix(filenameWithSuffix, fileSuffix)
	fmt.Println("fileNameOnly =", fileNameOnly)
	return fileNameOnly
}

//上传照片方法
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		err := renderHtml(w, "upload", nil)
		check(err)
		return
	}

	if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		if err != nil {
			http.Error(w, err.Error(),
				http.StatusInternalServerError)
			return
		}

		filename := h.Filename
		defer f.Close()
		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		check(err)
		defer t.Close()

		_, err = io.Copy(t, f)
		check(err)

		http.Redirect(w, r, "/view?id="+filename,
			http.StatusFound)
	}
}

//查看照片方法
func viewHandler(w http.ResponseWriter, r *http.Request) {
	imageId := r.FormValue("id")
	imagePath := UPLOAD_DIR + "/" + imageId

	if exists := isExists(imagePath); !exists {
		//http.NotFound(w, r)
		io.WriteString(w, "所访问的文件不存在！")
		return
	}

	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, imagePath)
	fmt.Println("view image :", imagePath)
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return os.IsExist(err)
}

//获取照片列表
func listHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("listHandler func")
	fileInfoArr, err := ioutil.ReadDir(UPLOAD_DIR)
	check(err)

	locals := make(map[string]interface{})
	images := []string{}

	for _, fileInfo := range fileInfoArr {

		images = append(images, fileInfo.Name())
	}
	locals["images"] = images

	err = renderHtml(w, "list", locals)
	check(err)
}

func renderHtml(w http.ResponseWriter, tmpl string,
	locals map[string]interface{}) (err error) {
	err = templates[tmpl].Execute(w, locals)
	return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

//返回闭包函数
func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(w, e.Error(), http.StatusInternalServerError)
				//自定义错误
				/*
					w.WriteeHeader(http.StatusInternalServerError)
					renderHtml(w, "error", e)
				*/
				//logging
				log.Println("WARN: panic in %v - %v", fn, e)
			}
		}()
		fn(w, r)
	}
}

func main() {
	mux := http.NewServeMux()
	staticDirHandler(mux, "/asserts/", "./public", 0)
	http.HandleFunc("/upload", safeHandler(uploadHandler))
	http.HandleFunc("/view", safeHandler(viewHandler))
	http.HandleFunc("/", safeHandler(listHandler))
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
