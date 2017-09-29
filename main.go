package main

import (
	"database/sql"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/kennygrant/sanitize"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	var v = Channel{}
	db, _ := sql.Open("sqlite3", "./main.db")
	defer db.Close()
	rows, _ := db.Query("SELECT guid, guid FROM guid")
	defer rows.Close()
	var gSclice []string
	for rows.Next() {
		var row string
		var guid string
		rows.Scan(&row, &guid)
		gSclice = append(gSclice, guid)
	}
	response, _ := http.Get(xmlUrl) //Get XML
	defer response.Body.Close()
	responseData, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(responseData, &v)
	for _, e := range reverseSclice(v.Items) {
		guid := e.Guid
		if stringInSlice(guid, gSclice) {
			println("Вывод")
			db.Exec("INSERT INTO guid(guid) VALUES (?)", guid)
			removeTags := sanitize.HTML(e.Description)                               //Remove html tags
			repTags := strings.NewReplacer("<br />", "\n", "&gt;", ">", "&nbsp", "") //Replace tags
			r := repTags.Replace(removeTags)
			message := url.QueryEscape("Title: " + e.Title + "\nАвтор: " + e.Author + "\nДата: " + e.PubDate + "\nСообщение: " + r)
			_, err := http.Get(apiRequest + message)
			if err != nil {
				errorMessage := url.QueryEscape(" Сообщение не может быть отправлено, так как его длина превышает максимальное количество символов или вызывает ошибку, посмотреть его можно по ссылке: https://fido.g0x.ru/rss?area://RU.FIDONET.TODAY")
				http.Get(apiRequest + errorMessage)
			}
			if len(gSclice) >= 30 { //Remove old guid values
				db.Exec("DELETE FROM guid WHERE rowid IN (SELECT rowid FROM guid ORDER BY rowid ASC LIMIT 1)")

			}
		}
	}
}

const xmlUrl = "https://fido.g0x.ru/rss?area://RU.FIDONET.DIGEST"
const apiRequest = "https://api.telegram.org/bot440564968:AAGj4IBd-byzsOeANzJhHDnsG4wkl4VuHxY/sendmessage?chat_id=-1001146113668&text="

type Rss struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}
type Channel struct {
	Items []Item `xml:"channel>item"`
}
type Item struct {
	Title       string `xml:"title"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Author      string `xml:"author"`
	Description string `xml:"description"`
}

func stringInSlice(a string, list []string) bool { //Check guid in database
	for _, b := range list {
		if b == a {
			return false
		}
	}
	return true
}
func reverseSclice(input []Item) []Item {
	if len(input) == 0 {
		return input
	}
	return append(reverseSclice(input[1:]), input[0])
}
