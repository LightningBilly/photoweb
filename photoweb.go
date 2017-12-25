package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"io/ioutil"
)

const (
	UPLOAD_DIR = "./uploads"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		io.WriteString(w, "<body>")
		io.WriteString(w, "<form method=\"POST\" action=\"/upload\" "+
			" enctype=\"multipart/form-data\" >"+
			"Choose an impage to upload: <input name=\"image\" type=\"file\" />"+
			"<input type=\"submit\" value=\"Upload\" />"+
			"</form>")

		io.WriteString(w, "</body>")

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
		if err != nil {
			http.Error(w, err.Error(),
				http.StatusInternalServerError)
			return
		}
		defer t.Close()

		if _, err := io.Copy(t, f); err != nil {
			http.Error(w, err.Error(),
				http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/view?id="+filename,
			http.StatusFound)
	}
}

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

func listHandler(w http.ResponseWriter, r*http.Request) {
	fileInfoArr, err := ioutil.ReadDir(UPLOAD_DIR)
	if err!=nil{
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	var listHtml string
	for _, fileInfo := range fileInfoArr {
		imgid := fileInfo.Name()
		fmt.Printf("%s\n", imgid)
		listHtml += "<li><a href = \"/view?id=" + imgid + "\">imgid</a></li>"
	}

	io.WriteString(w, "<body><ol>" + listHtml + "</ol></body>")
}

func main() {
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/view", viewHandler)
	http.HandleFunc("/", listHandler)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
