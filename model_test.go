package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"testing"
)

// テスト事にDBをリセットするため
func truncateTables() {
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		DELETE FROM albums;
		DELETE FROM pages;
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	if fileExists(dbFilePath) == true {
		if err := os.Remove(dbFilePath); err != nil {
			log.Fatal(err)
		}
	}
	if err := DBCreate(); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func TestAlbumCreateBySave(t *testing.T) {
	defer truncateTables()

	expect := "test title"
	m := &album{Title: expect}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	res, err := FindAlbumById(m.Id)
	if err != nil {
		t.Errorf("Save関数で作成したアルバムが見つかりませんでした.")
	}
	if res.Title != expect {
		t.Errorf("Save関数で作成したアルバムのタイトルが保存したものと異なります.Expect:%v, Actual:%v", res.Title, expect)
	}
}

func TestAlbumCreateValidate(t *testing.T) {
	defer truncateTables()

	tooLongTitle := "123456789012345678901234567890123"
	m := &album{Title: tooLongTitle}
	if err := m.Save(); err == nil {
		t.Errorf("32文字制限であるはずのタイトルに33文字での登録が行われました")
	}
}

func TestAlbumUpdateBySave(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	res, err := FindAlbumById(m.Id)
	if err != nil {
		t.Errorf("Save関数で作成したアルバムが見つかりませんでした.")
	}

	expect := "on update"

	res.Title = expect
	if err := res.Save(); err != nil {
		t.Fatal(err)
	}
	res, err = FindAlbumById(m.Id)
	if err != nil {
		t.Errorf("Save関数で作成したアルバムが見つかりませんでした.")
	}
	if res.Title != expect {
		t.Errorf("Save関数でTitleがアップデートされていません.")
	}
}

func TestAlbumFind(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	m = &album{Title: "test title2"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	// 無条件
	res, err := FindAlbum("")
	if err != nil {
		t.Fatal("FindAlbumでエラーが発生しました.", err)
	}
	expect := 2
	actual := len(res)
	if expect != actual {
		t.Errorf("FindAlbumで取得したデータ数が登録したデータ数と合いませんでした.Expect: %v, Actual: %v", expect, actual)
	}

	// 条件あり(1件絞り込み)
	res, err = FindAlbum("1")
	if err != nil {
		t.Fatal("FindAlbumでエラーが発生しました.", err)
	}
	expect = 1
	actual = len(res)
	if expect != actual {
		t.Errorf("FindAlbumで取得したデータ数が登録したデータ数と合いませんでした.Expect: %v, Actual: %v", expect, actual)
	}

	// 条件あり(該当なし)
	res, err = FindAlbum("18")
	if err != nil {
		t.Fatal("FindAlbumでエラーが発生しました.", err)
	}
	expect = 0
	actual = len(res)
	if expect != actual {
		t.Errorf("FindAlbumで取得したデータ数が登録したデータ数と合いませんでした.Expect: %v, Actual: %v", expect, actual)
	}
}

func TestRemoveAlbum(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	m = &album{Title: "test title2"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	if err := m.Remove(); err != nil {
		t.Fatal("Removeでアルバムの削除に失敗しました", err)
	}

	res, err := FindAlbum("")
	if err != nil {
		t.Fatal(err)
	}
	expect := 1
	actual := len(res)
	if expect != actual {
		t.Errorf("Removeで削除した後のFindAlbumで取得したデータ数が登録したデータ数と合いませんでした.Expect: %v, Actual: %v", expect, actual)
	}
}

func TestPageCreateBySave(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	p := &page{AlbumId: m.Id, Title: "test page title", Description: "desc"}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	findp, err := FindPageById(p.Id)
	if err != nil {
		t.Fatal(err)
	}

	if p.Id != findp.Id {
		t.Errorf("Page.Saveで作成したページとFindPageByIdで取得したページのIDが一致しませんでした.Expect: %v, Actual: %v", p.Id, findp.Id)
	}
	if p.AlbumId != findp.AlbumId {
		t.Errorf("Page.Saveで作成したページとFindPageByIdで取得したページのAlbumIDが一致しませんでした.Expect: %v, Actual: %v", p.AlbumId, findp.AlbumId)
	}
	if p.Title != findp.Title {
		t.Errorf("Page.Saveで作成したページとFindPageByIdで取得したページのTitleが一致しませんでした.Expect: %v, Actual: %v", p.Title, findp.Title)
	}
	if p.Description != findp.Description {
		t.Errorf("Page.Saveで作成したページとFindPageByIdで取得したページのDescriptionが一致しませんでした.Expect: %v, Actual: %v", p.Description, findp.Description)
	}
}

func TestPageUpdateBySave(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	p := &page{AlbumId: m.Id, Title: "test page title", Description: "desc"}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	findp, err := FindPageById(p.Id)
	if err != nil {
		t.Fatal(err)
	}

	expect_title := "updated title"
	expect_desc := "updated desc"

	p.Title = expect_title
	p.Description = expect_desc
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	findp, err = FindPageById(p.Id)
	if err != nil {
		t.Fatal(err)
	}

	if findp.Title != expect_title {
		t.Errorf("Page.Saveで更新したページとFindPageByIdで取得したページのTitleが一致しませんでした.Expect: %v, Actual: %v", p.Title, findp.Title)
	}
	if findp.Description != expect_desc {
		t.Errorf("Page.Saveで更新したページとFindPageByIdで取得したページのDescriptionが一致しませんでした.Expect: %v, Actual: %v", p.Description, findp.Description)
	}
}

func TestPageCreateValidate(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	p := &page{Title: "123456789012345678901234567890123"}
	if err := p.Save(); err == nil {
		t.Errorf("page.Saveで32文字制限のはずのタイトルに33文字登録することができました")
	}
	p = &page{Title: ""}
	if err := p.Save(); err == nil {
		t.Errorf("page.Saveで必須入力のはずのタイトルに空文字での登録を行うことができました")
	}

	s := "0123456789"
	overstr := "s"
	for i := 0; i < 100; i++ {
		overstr += s
	}
	p = &page{Title: "test", Description: overstr}
	if err := p.Save(); err == nil {
		t.Errorf("page.Saveで1000文字までの説明分に1001文字の登録を行うことができました")
	}
}

func TestPageRemove(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	p := &page{AlbumId: m.Id, Title: "test page title", Description: "desc"}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	findp, err := FindPageById(p.Id)
	if err != nil {
		t.Fatal(err)
	}

	if err := findp.Remove(); err != nil {
		t.Fatal("登録済みのpageの削除に失敗しました.", err)
	}

	if bad_find, err := FindPageById(p.Id); err == nil {
		t.Errorf("削除したはずのpageデータが残っています.Expect: nil, Actual:%v", bad_find)
	}
}

func TestFindPage(t *testing.T) {
	defer truncateTables()

	m := &album{Title: "test title1"}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	p1 := &page{AlbumId: m.Id, Title: "test page title1", Description: "desc1"}
	if err := p1.Save(); err != nil {
		t.Fatal(err)
	}
	p2 := &page{AlbumId: m.Id, Title: "test page title2", Description: "desc2"}
	if err := p2.Save(); err != nil {
		t.Fatal(err)
	}

	pages, err := FindPageByAlbumId(m.Id)
	if err != nil {
		t.Fatal("FindPageByAlbumIdでエラーが発生しました.", err)
	}

	if len(pages) != 2 {
		t.Errorf("FindPageByAlbumIdで検索されたデータ数が登録したデータ数と異なります.")
	}
	if pages[0].Title != p1.Title && pages[1].Title != p1.Title {
		t.Errorf("FindPageByAlbumIdで検索されたTitleが登録したデータと異なります.")
	}
	if pages[0].Description != p1.Description && pages[1].Description != p1.Description {
		t.Errorf("FindPageByAlbumIdで検索されたDescriptionが登録したデータと異なります.")
	}
}

func TestFindPageListData(t *testing.T) {
	defer truncateTables()

	expect_title := "test title1"
	m := &album{Title: expect_title}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	p1 := &page{AlbumId: m.Id, Title: "test page title1", Description: "desc1"}
	if err := p1.Save(); err != nil {
		t.Fatal(err)
	}
	p2 := &page{AlbumId: m.Id, Title: "test page title2", Description: "desc2"}
	if err := p2.Save(); err != nil {
		t.Fatal(err)
	}

	awp, err := FindPageListData(m.Id)
	if err != nil {
		t.Fatal("FindPageListDataでエラーが発生しました.", err)
	}

	if awp.Album.Title != expect_title {
		t.Errorf("FindPageListDataで取得したアルバム名が登録したアルバム名と一致しませんでした.Expect:%v, Actual:%v", expect_title, awp.Album.Title)
	}

	if len(awp.Pages) != 2 {
		t.Errorf("FindPageListDataで検索されたデータ数が登録したデータ数と異なります.")
	}

	if awp.Pages[0].Title != p1.Title && awp.Pages[1].Title != p1.Title {
		t.Errorf("FindPageListDataで検索されたTitleが登録したデータと異なります.")
	}
	if awp.Pages[0].Description != p1.Description && awp.Pages[1].Description != p1.Description {
		t.Errorf("FindPageListDataで検索されたDescriptionが登録したデータと異なります.")
	}
}
