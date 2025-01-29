package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

const WIKIURL = "https://www.pcgamingwiki.com/w/api.php"

func main() {
	// baseurl, err := url.Parse(WIKIURL)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// params := baseurl.Query()
	// params.Add("action", "query")
	// params.Add("list", "categorymembers")
	// params.Add("cmtitle", "Category:Games")
	// params.Add("cmlimit", "50")
	// params.Add("format", "json")

	// baseurl.RawQuery = params.Encode()

	// response, err := http.Get(baseurl.String())
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer response.Body.Close()

	// buf := bytes.NewBuffer(nil)
	// io.Copy(buf, response.Body)
	// // log.Println(string(buf.Bytes()))

	// var data map[string]interface{}
	// err = json.Unmarshal(buf.Bytes(), &data)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	GetGameSave(573)
	// params.Add("cmcontinue", data["continue"].(map[string]interface{})["cmcontinue"].(string))
	// params.Add("continue", data["continue"].(map[string]interface{})["continue"].(string))

	// baseurl.RawQuery = params.Encode()
	// secondResponse, err := http.Get(baseurl.String())
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer secondResponse.Body.Close()

	// buf = bytes.NewBuffer(nil)
	// io.Copy(buf, secondResponse.Body)
	// log.Println(string(buf.Bytes()))
}

func GetGameSave(id float64) {
	baseurl, err := url.Parse(WIKIURL)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)
	// log.Println(string(buf.Bytes()))

	var sectionsResponse map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &sectionsResponse)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}

	buf = bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)

	var gameSavesResponse map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &gameSavesResponse)
	if err != nil {
		log.Fatalln(err)
	}

	parse = gameSavesResponse["parse"].(map[string]interface{})
	wikitext := parse["wikitext"].(map[string]interface{})
	text := wikitext["*"].(string)

	regex, err := regexp.Compile("{{Game data\\/saves\\|(.*?)\\|(.*?)}}\\n")
	if err != nil {
		log.Fatalln(err)
	}

	matches := regex.FindAllStringSubmatch(text, -1)
	log.Println(matches)
}
