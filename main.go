package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	ExpDuration = 10
)

type RequestInfo struct {
	RequestTime time.Time
	Expiration  time.Time
	Request     string
}

var (
	MainChan     = make(chan string)
	RequestsList = make(map[string]RequestInfo)
	Gmu          sync.Mutex
	GRmu         sync.RWMutex
)

func GWork() {
	for {
		if request, ok := <-MainChan; ok {
			timeNow := time.Now()
			Gmu.Lock()
			RequestsList[request] = RequestInfo{
				RequestTime: timeNow,
				Expiration:  timeNow.Add(ExpDuration * time.Minute),
				Request:     request,
			}
			Gmu.Unlock()
		}
	}
}

func CheckExp() {

	ticker := time.Tick(20 * time.Second)

	for {
		if _, ok := <-ticker; ok {
			GRmu.RLock()
			for key, value := range RequestsList {
				if time.Now().After(value.Expiration) {
					RequestsList[key] = RequestInfo{}
				}
			}
			GRmu.RUnlock()
		}
	}
}

func Monitor() {
	for tick := range time.Tick(10 * time.Second) {
		GRmu.RLock()
		fmt.Println(tick)
		for key, value := range RequestsList {
			fmt.Println(key, value)
		}
		fmt.Println("---------------------------")
		GRmu.RUnlock()
	}
}

func main() {
	go GWork()
	go CheckExp()
	go Monitor()

	r := gin.Default()
	r.GET("/request/:request", Request)

	r.Run(":8080")
}

func Request(c *gin.Context) {

	request := c.Param("request")

	MainChan <- request

	c.JSON(http.StatusOK, "OK")
}
