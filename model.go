package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"unicode/utf8"
)

const dbFilePath = "album.db"

func DBCreate() error {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE "albums" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"title" VARCHAR(32)
		);
		CREATE TABLE "pages" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"album_id" INTEGER NOT NULL,
			"title" VARCHAR(128) NOT NULL,
			"description" VARCHAR(1024) NOT NULL,
			"filepath" VARCHAR(1024)
		);
	`)
	if err != nil {
		return err
	}
	return nil
}

type album struct {
	Id    int64
	Title string
}

func FindAlbum(title_cond string) ([]*album, error) {
	query := `
		SELECT
			id    AS id,
			title AS title
		FROM
			albums
			`
	option := `
		WHERE
			title LIKE ?
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	if len(title_cond) > 0 {
		query += option
	}
	rows, err := db.Query(query, "%"+title_cond+"%")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make([]*album, 0)
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		m := &album{Id: id, Title: title}
		ret = append(ret, m)
	}

	return ret, nil
}

func FindAlbumById(id int64) (*album, error) {
	query := `
		SELECT
			title AS title
		FROM
			albums
		WHERE
			id = ?
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}
	var title string
	err = db.QueryRow(query, id).Scan(&title)
	if err != nil {
		return nil, err
	}

	m := &album{Id: id, Title: title}

	return m, nil
}

func (m *album) create() error {
	query := `
		INSERT INTO albums (title) values(?)
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	res, err := db.Exec(query, m.Title)
	if err != nil {
		return err
	}
	m.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (m *album) update() error {
	query := `
		UPDATE
			albums
		SET
			title = ?
		WHERE
			id = ?
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	_, err = db.Exec(query, m.Title, m.Id)
	if err != nil {
		return err
	}

	return nil
}

func (m *album) Validate() error {
	if utf8.RuneCountInString(m.Title) > 32 {
		return errors.New("Title is too long")
	}
	return nil
}

func (m *album) Save() error {
	_, err := FindAlbumById(m.Id)
	if err := m.Validate(); err != nil {
		return err
	}
	if err == nil {
		return m.update()
	} else {
		return m.create()
	}
}

func (m *album) Remove() error {
	query := `
		DELETE
		FROM
			albums
		WHERE
			id = ?
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	_, err = db.Exec(query, m.Id)
	if err != nil {
		return err
	}

	return nil
}

type page struct {
	Id          int64
	AlbumId     int64
	Title       string
	Description string
	MoviePath   string
}

func FindPageByAlbumId(albumId int64) ([]*page, error) {
	query := `
		SELECT
			page.id    AS id,
			page.album_id AS album_id,
			page.title AS title,
			page.description AS description,
			page.filepath AS filepath
		FROM
			pages page
		WHERE
			page.album_id = ?
			`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, albumId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make([]*page, 0)
	for rows.Next() {
		var id int64
		var albumId int64
		var title string
		var description string
		var filepath sql.NullString
		if err := rows.Scan(&id, &albumId, &title, &description, &filepath); err != nil {
			return nil, err
		}
		filepathstr := ""
		if filepath.Valid {
			filepathstr = filepath.String
		}
		m := &page{Id: id, AlbumId: albumId, Title: title, Description: description, MoviePath: filepathstr}
		ret = append(ret, m)
	}

	return ret, nil
}

func FindPageById(pageId int64) (*page, error) {
	query := `
		SELECT
			page.album_id AS album_id,
			page.title AS title,
			page.description AS description,
			page.filepath AS filepath
		FROM
			pages page
		WHERE
			page.id = ?
			`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, pageId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albumId int64
	var title string
	var description string
	var filepath sql.NullString
	err = db.QueryRow(query, pageId).Scan(&albumId, &title, &description, &filepath)
	if err != nil {
		return nil, err
	}
	filepathstr := ""
	if filepath.Valid {
		filepathstr = filepath.String
	}
	m := &page{Id: pageId, AlbumId: albumId, Title: title, Description: description, MoviePath: filepathstr}
	return m, nil
}

func (m *page) create() error {
	query := `
		INSERT INTO pages (album_id, title, description, filepath) values(?, ?, ?, ?)
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	res, err := db.Exec(query, m.AlbumId, m.Title, m.Description, m.MoviePath)
	if err != nil {
		return err
	}
	m.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (m *page) update() error {
	query := `
		UPDATE
			pages
		SET
			album_id = ?,
			title = ?,
			description = ?
		WHERE
			id = ?
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	_, err = db.Exec(query, m.AlbumId, m.Title, m.Description, m.Id)
	if err != nil {
		return err
	}

	return nil
}

func (m *page) Validate() error {
	if utf8.RuneCountInString(m.Title) > 32 {
		return errors.New("title is too long")
	}
	if utf8.RuneCountInString(m.Title) == 0 {
		return errors.New("title is nesecery.")
	}
	if utf8.RuneCountInString(m.Description) > 1000 {
		return errors.New("description is too long")
	}
	return nil
}

func (m *page) Save() error {
	if err := m.Validate(); err != nil {
		return err
	}
	_, err := FindPageById(m.Id)
	if err == nil {
		return m.update()
	} else {
		return m.create()
	}
}

func (m *page) Remove() error {
	query := `
		DELETE
		FROM
			pages
		WHERE
			id = ?
	`
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return err
	}
	_, err = db.Exec(query, m.Id)
	if err != nil {
		return err
	}
	return nil
}

type pageListData struct {
	Album      *album
	Pages      []*page
	SelectPage *page
}

func FindPageListData(albumId int64) (*pageListData, error) {
	album, err := FindAlbumById(albumId)
	if err != nil {
		return nil, err
	}
	pages, err := FindPageByAlbumId(album.Id)
	if err != nil {
		return nil, err
	}
	pld := &pageListData{Album: album, Pages: pages}
	if len(pages) > 0 {
		pld.SelectPage = pages[0]
	}
	return pld, nil
}

type pageEditData struct {
	Album      *album
	SelectPage *page
}

func FindPageEditData(albumId int64, pageId int64) (*pageEditData, error) {
	album, err := FindAlbumById(albumId)
	if err != nil {
		return nil, err
	}
	page, err := FindPageById(pageId)
	ped := &pageEditData{Album: album, SelectPage: page}
	return ped, nil
}
