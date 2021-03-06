package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// TwitterCreds contains API credentials to access Twitter API
// For more information check twitter API documentation
type TwitterCreds struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

// StockData is what we read from borsa.com, at least what we need
type StockData struct {
	PreviousClose float64 `json:"previous_closing"`
	Latest        float64 `json:"latest"`
	ChangeRate    float64 `json:"change_rate"`
}

// read api keys from environmental variables
func readCreds() TwitterCreds {
	config := TwitterCreds{}
	key := os.Getenv("CONSUMERKEY")
	if key == "" {
		panic("empty consumer key")
	}
	config.ConsumerKey = key
	key = os.Getenv("CONSUMERSECRET")
	if key == "" {
		panic("empty consumer secret")
	}
	config.ConsumerSecret = key
	key = os.Getenv("ACCESSTOKEN")
	if key == "" {
		panic("empty access token")
	}
	config.AccessToken = key
	key = os.Getenv("ACCESSSECRET")
	if key == "" {
		panic("empty access seceret")
	}
	config.AccessSecret = key

	return config
}

func main() {
	// read credentials
	config := readCreds()

	// create twitter client
	oauthCfg := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.AccessToken, config.AccessSecret)

	httpClient := oauthCfg.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)
	run(client)
}

func isWeekDay(t time.Time) bool {
	if t.Weekday() != time.Saturday && t.Weekday() != time.Sunday {
		return true
	}
	return false
}

func run(client *twitter.Client) {
	ticker := time.NewTicker(1 * time.Hour)
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		log.Printf("cannot find location database. Please install tzdata: %v", err)
		loc, _ = time.LoadLocation("UTC")
	}
	for {
		t := time.Now().In(loc)
		// Turkey is UTC+3. Markets close at 17.
		// Run it at the end of the each work day
		if t.Hour() == 17 && isWeekDay(t) {
			tweet(client)
		}
		<-ticker.C
	}
}

func tweet(client *twitter.Client) {
	// get stock market details
	resp, _ := http.Get("https://www.doviz.com/api/v1/indexes/XU100/latest")

	payload, _ := ioutil.ReadAll(resp.Body)
	var data StockData
	err := json.Unmarshal(payload, &data)
	if err != nil {
		log.Println("wrong data: ", err.Error())
		return
	}

	opening := strconv.FormatFloat(data.PreviousClose, 'f', 3, 64)
	closing := strconv.FormatFloat(data.Latest, 'f', 3, 64)

	// tweet
	var result string
	if data.Latest > data.PreviousClose {
		result = fmt.Sprintf("sıçmadı 😎\nBIST100 %%%.2f artışla kapandı.", data.ChangeRate)
	} else {
		result = fmt.Sprintf("sıçtı 🤬\nBIST100 %%%.2f düşüşle kapandı.", -data.ChangeRate)
	}
	status := fmt.Sprintf("%s\nAçılış: %s\nKapanış: %s", result, opening, closing)
	fmt.Println(status)
	_, res, err := client.Statuses.Update(status, nil)
	if err != nil {
		log.Println(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(res.Body)
		log.Printf("Twitter returned %d - %s", res.StatusCode, string(data))
	}
}
