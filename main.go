package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const moviesRoot = "movies"

var ViewTemplatesMap map[string]*template.Template

func init() {
	// テンプレートを事前パース
	ViewTemplatesMap = make(map[string]*template.Template)
	ViewTemplatesMap["album_list"] = template.Must(template.ParseFiles("view/album_list.html"))
	ViewTemplatesMap["page_list"] = template.Must(template.ParseFiles("view/page_list.html"))
	ViewTemplatesMap["page_edit"] = template.Must(template.ParseFiles("view/page_edit.html"))
}

/*
 * パース済みテンプレートを実行.
 * エラー時にはInternal Server Errorを出す.
 */
func execTemplate(w http.ResponseWriter, templateName string, obj interface{}) error {
	t := ViewTemplatesMap[templateName]
	if err := t.Execute(w, obj); err != nil {
		// エラー時はInternalServerError
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func get_albums(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	q := r.FormValue("q")
	albums, err := FindAlbum(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	execTemplate(w, "album_list", albums)
}

func get_album(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	id_str := r.FormValue("album_id")
	id, err := strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pld, err := FindPageListData(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id_str = r.FormValue("page_id")
	if id_str != "" {
		id, err = strconv.ParseInt(id_str, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pld.SelectPage, err = FindPageById(id)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	execTemplate(w, "page_list", pld)
}

func add_album(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	name := r.PostFormValue("album_name")
	album := &album{Title: name}
	if err := album.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pld, err := FindPageListData(album.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	execTemplate(w, "page_list", pld)
}

func delete_album(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	id_str := r.FormValue("album_id")
	id, err := strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	album, err := FindAlbumById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = album.Remove()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	albums, err := FindAlbum("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	execTemplate(w, "album_list", albums)
}

func new_page(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	id_str := r.FormValue("album_id")
	id, err := strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	album, err := FindAlbumById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ped := &pageEditData{Album: album, SelectPage: nil}
	execTemplate(w, "page_edit", ped)
}

func filesave(file io.Reader, name string) (path string, err error) {
	p := filepath.Join(moviesRoot, name+".mp4")
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(f, file); err != nil {
		return "", err
	}
	return name + ".mp4", nil
}

func randStr() string {
	var n uint64
	binary.Read(rand.Reader, binary.LittleEndian, &n)
	return strconv.FormatUint(n, 36)
}

func save_page(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	id_str := r.FormValue("album_id")
	album_id, err := strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page_id_str := r.FormValue("page_id")
	var page_id int64
	if page_id_str != "" {
		page_id, err = strconv.ParseInt(page_id_str, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	title := r.FormValue("title")
	desc := r.FormValue("description")

	file, _, err := r.FormFile("video")
	filepath := ""
	if err == nil {
		if filepath, err = filesave(file, randStr()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	p := &page{AlbumId: album_id, Title: title, Description: desc, MoviePath: filepath}
	if page_id != 0 {
		p.Id = page_id
	}
	if err := p.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pld, err := FindPageListData(album_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pld.SelectPage = pld.Pages[0]
	execTemplate(w, "page_list", pld)
}

func edit_page(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	id_str := r.FormValue("album_id")
	id, err := strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	album, err := FindAlbumById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id_str = r.FormValue("page_id")
	id, err = strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page, err := FindPageById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ped := &pageEditData{Album: album, SelectPage: page}
	execTemplate(w, "page_edit", ped)
}

func delete_page(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	id_str := r.FormValue("page_id")
	id, err := strconv.ParseInt(id_str, 10, 64)
	page, err := FindPageById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := page.Remove(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id_str = r.FormValue("album_id")
	id, err = strconv.ParseInt(id_str, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pld, err := FindPageListData(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	execTemplate(w, "page_list", pld)
}

func main() {

	port := flag.Int("p", 9000, "accept port number.")
	init := flag.Bool("i", false, "Initialize DB and Data directories.")
	flag.Parse()

	port_no := "9000"
	if *port != 0 {
		port_no = strconv.Itoa(*port)
	}

	if *init {
		if fileExists("album.db") == false {
			err := DBCreate()
			if err != nil {
				fmt.Println("On error occurred in init: ", err.Error())
			}
		}
		if fileExists(moviesRoot) == false {
			if err := os.Mkdir(moviesRoot, 0777); err != nil {
				panic(err)
			}
		}
		return
	}

	http.HandleFunc("/get_albums", get_albums)
	http.HandleFunc("/add_album", add_album)
	http.HandleFunc("/get_album", get_album)
	http.HandleFunc("/delete_album", delete_album)

	http.HandleFunc("/new_page", new_page)
	http.HandleFunc("/edit_page", edit_page)
	http.HandleFunc("/save_page", save_page)
	http.HandleFunc("/delete_page", delete_page)

	http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("assets"))))
	http.Handle("/movies/", http.StripPrefix("/movies", http.FileServer(http.Dir("movies"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		get_albums(w, r)
	})

	http.ListenAndServe(":"+port_no, nil)
}
