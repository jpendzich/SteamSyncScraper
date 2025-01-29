package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const WIKIURL = "https://www.pcgamingwiki.com/w/api.php"

func main() {
	baseurl, err := url.Parse(WIKIURL)
	if err != nil {
		log.Fatalln(err)
	}

	params := baseurl.Query()
	params.Add("action", "query")
	params.Add("list", "categorymembers")
	params.Add("cmtitle", "Category:Games")
	params.Add("cmlimit", "50")
	params.Add("format", "json")

	baseurl.RawQuery = params.Encode()

	response, err := http.Get(baseurl.String())
	if err != nil {
		log.Fatalln(err)
	}
	defer response.Body.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, response.Body)
	// log.Println(string(buf.Bytes()))

	var data map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &data)
	if err != nil {
		log.Fatalln(err)
	}
	GetGameSave(data["query"].(map[string]interface{})["categorymembers"].([]interface{})[0].(map[string]interface{})["pageid"].(float64))
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
	params.Add("prop", "wikitext")
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
	log.Println(string(buf.Bytes()))
}
