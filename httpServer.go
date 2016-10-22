package dht

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var chttp = http.NewServeMux()

func StartServer() {

	chttp.Handle("/", http.FileServer(http.Dir("./src/dht/server/")))

	http.HandleFunc("/", HomeHandler) // homepage
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/delete", deleteHandler)

	http.ListenAndServe(":8080", nil)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	chttp.ServeHTTP(w, r)
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/store")
	http.Redirect(w, r, url[0], 303)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	f, err := os.OpenFile("/tmp/"+handler.Filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	content, _ := ioutil.ReadFile(handler.Filename)
	storeFile(handler.Filename, content)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	getFile(r.FormValue("value"))
	// if !err {
	// 	fmt.Println("the content of the file is:", string(content))
	// } else {
	// 	fmt.Println("the file does not exists!")
	// }
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	storeHandler(w, r)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	deleteFile(r.FormValue("value"))
}
