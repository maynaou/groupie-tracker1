package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type API struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	Relations    string   `json:"relations"`
}

type Relation struct {
	DatesLocations map[string][]string `json:"DatesLocations"`
}

var (
	templates = template.Must(template.ParseFiles("templates/home.html", "templates/details.html"))
	ApiObject []API
)

func detailsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 || id > len(ApiObject) {
		http.NotFound(w, r)
		return
	}

	artist := ApiObject[id-1]
	fmt.Println(artist)
	resp, err := http.Get(artist.Relations)
	if err != nil {
		http.Error(w, "Error fetching relations", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	var relations Relation
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading relations data", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, &relations)
	if err != nil {
		http.Error(w, "Error decoding relations JSON", http.StatusInternalServerError)
		return
	}

	templates.ExecuteTemplate(w, "details.html", map[string]interface{}{
		"Artist":    artist,
		"Relations": relations,
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading data", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, &ApiObject)
	if err != nil {
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	templates.ExecuteTemplate(w, "home.html", ApiObject)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/artist", detailsHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
