package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const WIKIURL = "https://www.pcgamingwiki.com/w/api.php"
const CMLIMIT = 50

const (
	dropGamesTable   = `DROP TABLE IF EXISTS Saves;`
	createGamesTable = `CREATE TABLE Saves (
		PageId INTEGER PRIMARY KEY,
		Game TEXT,
		SaveLocation TEXT
	);`
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "saveLocations.db")
	if err != nil {

	}
	defer db.Close()

	_, err = db.Exec(dropGamesTable)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.Exec(createGamesTable)
	if err != nil {
		log.Fatalln(err)
	}

	baseurl, err := url.Parse(WIKIURL)
	if err != nil {
		log.Fatalln(err)
	}

	params := baseurl.Query()
	params.Add("action", "query")
	params.Add("list", "categorymembers")
	params.Add("cmtitle", "Category:Games")
	params.Add("cmlimit", fmt.Sprintf("%d", CMLIMIT))
	params.Add("format", "json")

	baseurl.RawQuery = params.Encode()

	continuePresent := true
	for continuePresent {
		log.Printf("new batch of %d\n", CMLIMIT)
		response, err := http.Get(baseurl.String())
		if err != nil {
			log.Printf("Error: %s | Trying again\n", err)
			continue
		}
		defer response.Body.Close()

		buf := bytes.NewBuffer(nil)
		io.Copy(buf, response.Body)
		// log.Println(string(buf.Bytes()))

		var data map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &data)
		if err != nil {
			log.Printf("Error: %s | Trying again\n", err)
			continue
		}
		query := data["query"].(map[string]interface{})
		catergorymembers := query["categorymembers"].([]interface{})

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(catergorymembers))
		for i := 0; i < len(catergorymembers); i++ {
			member := catergorymembers[i].(map[string]interface{})
			go saveGameSaveById(member["pageid"].(float64), member["title"].(string), &waitGroup) // use goroutine here to speed up query
		}

		waitGroup.Wait()

		if continueJson, ok := data["continue"].(map[string]interface{}); ok {
			params.Set("cmcontinue", continueJson["cmcontinue"].(string))
			params.Set("continue", continueJson["continue"].(string))
			baseurl.RawQuery = params.Encode()
		} else {
			break
		}
	}
}

func saveGameSaveById(id float64, game string, waitGroup *sync.WaitGroup) {
	baseurl, err := url.Parse(WIKIURL)
	if err != nil {
		waitGroup.Done()
		return
	}

	params := baseurl.Query()
	params.Add("action", "parse")
	params.Add("pageid", fmt.Sprintf("%.0f", id))
	params.Add("prop", "sections")
	params.Add("format", "json")
	params.Add("redirects", "1")
	baseurl.RawQuery = params.Encode()

	response, err := http.Get(baseurl.String())
	if err != nil {
		waitGroup.Done()
		return
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)
	// log.Println(string(buf.Bytes()))

	var sectionsResponse map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &sectionsResponse)
	if err != nil {
		waitGroup.Done()
		return
	}
	parse := sectionsResponse["parse"].(map[string]interface{})
	sections := parse["sections"].([]interface{})

	saveGameIndex := ""
	for i := 0; i < len(sections); i++ {
		section := sections[i].(map[string]interface{})
		if section["line"].(string) == "Save game data location" {
			saveGameIndex = section["index"].(string)
			break
		}
	}
	if saveGameIndex == "" {
		waitGroup.Done()
		return
	}

	params = url.Values{}
	params.Add("action", "parse")
	params.Add("pageid", fmt.Sprintf("%.0f", id))
	params.Add("prop", "wikitext")
	params.Add("section", saveGameIndex)
	params.Add("format", "json")
	params.Add("redirects", "1")
	baseurl.RawQuery = params.Encode()

	response, err = http.Get(baseurl.String())
	if err != nil {
		waitGroup.Done()
		return
	}

	buf = bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)

	var gameSavesResponse map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &gameSavesResponse)
	if err != nil {
		waitGroup.Done()
		return
	}

	parse = gameSavesResponse["parse"].(map[string]interface{})
	wikitext := parse["wikitext"].(map[string]interface{})
	text := wikitext["*"].(string)

	regex, err := regexp.Compile("{{Game data\\/saves\\|(.*?)\\|(.*?)}}\\n")
	if err != nil {
		waitGroup.Done()
		return
	}

	matches := regex.FindAllStringSubmatch(text, -1)
	saves := make(map[string]string)
	gameHasAnySaves := false
	for i := 0; i < len(matches); i++ {
		//only make an entry if the second capture group returned a non empty string
		if (len(matches[i]) >= 3) && (matches[i][2] != "") {
			saves[matches[i][1]] = matches[i][2]

			if !gameHasAnySaves {
				//only set when it wasn't previously true
				gameHasAnySaves = true
			}
		}
	}

	if gameHasAnySaves { //don't bother saving the game when it doesn't have any know save locations
		savesJson, err := json.Marshal(saves)
		if err != nil {
			waitGroup.Done()
			return
		}

		_, err = db.Exec("INSERT INTO Saves VALUES(?, ?, ?)", fmt.Sprintf("%.0f", id), game, string(savesJson))
		if err != nil {
			waitGroup.Done()
			return
		}
		log.Printf("saved %s with id %.0f", game, id)
	}
	waitGroup.Done()
	return
}
