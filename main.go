package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/gorilla/mux"
)

type Info struct {
	ip      string
	status  string
	traffic string
	time    int
}

var info Info

func main() {

	//AWS API
	var AKID string = os.Getenv("AKID")
	var SECRET string = os.Getenv("SECRET")
	var REGION string = os.Getenv("REGION")
	var PORT string = os.Getenv("PORT") //heroku auto deployment setting

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(REGION),
		Credentials: credentials.NewStaticCredentials(AKID, SECRET, ""),
	}))
	svc := lightsail.New(sess)

	info = Info{GetPublicIP(svc), GetStatus(svc), GetTotalNetworkPerMonth(svc), 0}

	//HTTP Server
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/ws", wsHandler)
	go echo()
	go func() {
		var i int = 0
		for range time.Tick(time.Second) {
			if i > 300 {
				i = 0
			}
			info.time = i
			if i == 255 {
				go func() {
					info.ip = GetPublicIP(svc)
					info.status = GetStatus(svc)
					info.traffic = GetTotalNetworkPerMonth(svc)
				}()
			}
			infoWriter(info)
			i++
		}
	}()
	log.Fatal(http.ListenAndServe(":"+PORT, router))
}

func GetConfig(filename string) map[string]interface{} {
	file, _ := ioutil.ReadFile(filename)
	var config map[string]interface{}
	json.Unmarshal(file, &config)
	return config
}
