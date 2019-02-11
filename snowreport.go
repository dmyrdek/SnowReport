package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	loggly "github.com/jamespearly/loggly"
	cron "gopkg.in/robfig/cron.v2"
)

type SnowReport struct {
	Resortid      float64 `json:"resortid"`
	Resortname    string  `json:"resortname"`
	Resortcountry string  `json:"resortcountry"`
	Newsnow_cm    float64 `json:"newsnow_cm"`
	Newsnow_in    float64 `json:"newsnow_in"`
	Lowersnow_cm  float64 `json:"lowersnow_cm"`
	Lowersnow_in  float64 `json:"lowersnow_in"`
	Uppersnow_cm  float64 `json:"uppersnow_cm"`
	Uppersnow_in  float64 `json:"uppersnow_in"`
	Pctopen       float64 `json:"pctopen"`
	Lastsnow      string  `json:"lastsnow"`
	Reportdate    string  `json:"reportdate"`
	Reporttime    string  `json:"reporttime"`
	Conditions    string  `json:"conditions"`
}

func main() {

	key, found := os.LookupEnv("LOGGLY_TOKEN")
	if !found {
		fmt.Println("Can't find variable")
	}
	fmt.Println(key)

	c := cron.New()
	c.AddFunc("@every 1m", getSnowReport)
	c.Start()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

}

func getSnowReport() {
	var tag string
	tag = "Kitzbuhel-Snow-Report"

	logglyClient := loggly.New(tag)

	url := "https://api.weatherunlocked.com/api/snowreport/..."

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("NewRequest: ", err)
		logglyLog := logglyClient.EchoSend("NewRequest: ", err.Error())
		fmt.Println("logglyLog:", logglyLog)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		logglyLog := logglyClient.EchoSend("Do: ", err.Error())
		fmt.Println("logglyLog:", logglyLog)
		return
	}

	defer resp.Body.Close()

	var record SnowReport

	currentTime := time.Now()

	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		log.Println(err)
		logglyLog := logglyClient.EchoSend(currentTime.Format("2006-01-02 15:04:05"), err.Error())
		fmt.Println("logglyError:", logglyLog)
	}
	logglyLog := logglyClient.EchoSend("info", floattostr(record.Newsnow_in))
	fmt.Println("info: ", logglyLog)
	fmt.Println("resortid      = ", record.Resortid)
	fmt.Println("resortname    = ", record.Resortname)
	fmt.Println("resortcountry = ", record.Resortcountry)
	fmt.Println("newsnow_cm    = ", record.Newsnow_cm)
	fmt.Println("newsnow_in    = ", record.Newsnow_in)
	fmt.Println("lowersnow_cm  = ", record.Lowersnow_cm)
	fmt.Println("lowersnow_in  = ", record.Lowersnow_in)
	fmt.Println("uppersnow_cm  = ", record.Uppersnow_cm)
	fmt.Println("uppersnow_in  = ", record.Uppersnow_in)
	fmt.Println("pctopen       = ", record.Pctopen)
	fmt.Println("lastsnow      = ", record.Lastsnow)
	fmt.Println("reportdate    = ", record.Reportdate)
	fmt.Println("reporttime    = ", record.Reporttime)
	fmt.Println("conditions    = ", record.Conditions)
	fmt.Println(" ")
}

func floattostr(input_num float64) string {

	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'g', 1, 64)
}
