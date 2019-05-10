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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

type Item struct {
	Reportdate string
	Newsnow_in float64
}

func main() {

	key, found := os.LookupEnv("LOGGLY_TOKEN")
	if !found {
		fmt.Println("Can't find variable")
	}
	fmt.Println(key)

	getSnowReport()
	c := cron.New()
	c.AddFunc("@every 1d", getSnowReport)
	c.Start()

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

}

func sendResponseToDynamoDB(snowreport SnowReport) {
	var tag string
	tag = "Kitzbuhel-Snow-Report"

	logglyClient := loggly.New(tag)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	if err != nil {
		log.Fatal("Item Failed: ", err)
		logglyLog := logglyClient.EchoSend("Item Failed: ", err.Error())
		fmt.Println("logglyLog:", logglyLog)
		return
	}

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	it := Item{
		snowreport.Reportdate,
		snowreport.Newsnow_in,
	}

	av, err := dynamodbattribute.MarshalMap(it)

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("SnowReport"),
	}

	_, err = svc.PutItem(input)

	if err != nil {
		//log.Fatal("Put Failed: ", err)
		fmt.Println(err)
		logglyLog := logglyClient.EchoSend("Put Failed: ", err.Error())
		fmt.Println("logglyLog:", logglyLog)
		return
	}

	fmt.Print("Succesfully addded all items to DynamoDB!")

}

func getSnowReport() {
	var tag string
	tag = "Kitzbuhel-Snow-Report"

	logglyClient := loggly.New(tag)

	url := "https://api.weatherunlocked.com/api/snowreport/222013?app_id=08c12f0a&app_key=13ae4e2cd3b974483ea0ac6903ac8cfc"

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
	sendResponseToDynamoDB(record)
}

func floattostr(input_num float64) string {

	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'g', 1, 64)
}
