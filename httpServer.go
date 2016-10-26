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
var circleJoined = false

func StartServer(port string) {

	chttp.Handle("/", http.FileServer(http.Dir("./src/dht/server/")))

	http.HandleFunc("/", HomeHandler) // homepage
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/join", joinHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/leave", leaveHandler)

	http.ListenAndServe(":"+port, nil)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	chttp.ServeHTTP(w, r)
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	_, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, notFound := (getFile(handler.Filename))
	if !notFound { //File already exist - stop upload
		w.Write([]byte("FileExists"))
		return
	}
	fmt.Println("Storing", handler.Filename)
	store(w, r)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("value")
	content, err := getFile(fileName)
	if !err {
		//fmt.Println("the content of the file is:", string(content))
	} else {
		fmt.Println("The file does not exists!")
	}
	w.Write(content)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Updating")
	r.ParseForm()
	_, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, notFound := (getFile(handler.Filename))
	if notFound { //File does not  exist - stop update
		//According to the frontend, this should not happen
		w.Write([]byte("NotFileExists"))
		return
	}
	store(w, r)
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

func joinHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	port := r.URL.Query().Get("port")
	fmt.Println("ip:", ip, "port:", port)
	tojoin := StringToContact(ip + "-" + port + "-")
	circleJoined = joinRing(contact, &tojoin)
	w.Write([]byte(strconv.FormatBool(circleJoined)))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	response := contact.NodeId + "/" + strconv.FormatBool(circleJoined)
	w.Write([]byte(response))
}

func leaveHandler(w http.ResponseWriter, r *http.Request) {
	leaveRing()
}

func store(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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

	content, _ := ioutil.ReadFile("/tmp/" + handler.Filename)
	storeFile(handler.Filename, content)
	w.Write([]byte("OK"))
}

func setJoined() {
	circleJoined = true
}
