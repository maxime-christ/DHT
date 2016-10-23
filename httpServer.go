package dht

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

var chttp = http.NewServeMux()

func StartServer() {

	chttp.Handle("/", http.FileServer(http.Dir("./src/dht/server/")))

	http.HandleFunc("/", HomeHandler) // homepage
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/search", searchHandler)

	http.ListenAndServe(":8080", nil)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	chttp.ServeHTTP(w, r)
}

func storeHandler(w http.ResponseWriter, r *http.Request) {

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
	defer os.Remove("/tmp/" + handler.Filename)

	io.Copy(f, file)

	// content := make([]byte, 3<<20)
	// n, err := f.Read(content)
	// fmt.Println("read", n, "bytes, err is", err, "content is :", string(content))

	content, _ := ioutil.ReadFile("/tmp/" + handler.Filename)
	storeFile(handler.Filename, content)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("value")
	content, err := getFile(fileName)
	if !err {
		//fmt.Println("the content of the file is:", string(content))
	} else {
		fmt.Println("the file does not exists!")
	}
	w.Write(content)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Updating")
	storeHandler(w, r)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("value")
	fmt.Println("delete file : ", fileName)
	deleteFile(fileName)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("value")
	_, notFound := getFile(fileName)
	w.Write([]byte(strconv.FormatBool(notFound)))
}
