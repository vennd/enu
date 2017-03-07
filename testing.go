package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/vennd/enu/enulib"
	"github.com/vennd/enu/log"
)

var isInit = false
var apiKey = "71625888dc50d8915b871912aa6bbdce67fd1ed77d409ef1cf0726c6d9d7cf16"
var apiSecret = "a06d8cfa8692973c755b3b7321a8af7de448ec56dcfe3739716f5fa11187e4ac"
var baseURL = "http://localhost:8081"
var once sync.Once
var ready sync.Mutex

var done bool

// Creates a local server to serve client tests
func InitTesting(c chan bool) {
	if isInit == true {
		c <- true
		return
	}

	//	log.Printf("Initilising...")
	time.Sleep(time.Duration(5000) * time.Millisecond) // introduce start up time to test race condition on a fast machine

	router := NewRouter()
	isInit = true
	c <- true

	log.Println("Enu Unit Test API server started")
	log.Println(http.ListenAndServe("localhost:8081", router).Error())
}

func DoEnuAPITesting(method string, url string, postData []byte) ([]byte, int, error) {
	c := make(chan bool, 1)
	go InitTesting(c)

	select {
	case done = <-c:
	case <-time.After(time.Second * 10):
		log.Println("Timed out whilst waiting for unit testing server to initialise")
		os.Exit(1)
	}

	if done == false {
		log.Println("Error initialising unit test server")
		os.Exit(1)
	}

	if method != "POST" && method != "GET" {
		return nil, -1000, errors.New("DoEnuAPI must be called for a POST or GET method only")
	}
	postDataJson := string(postData)

	log.Printf("Posting: %s", postDataJson)

	// Set headers
	req, err := http.NewRequest(method, url, bytes.NewBufferString(postDataJson))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accessKey", apiKey)
	req.Header.Set("signature", enulib.ComputeHmac512(postData, apiSecret))

	// Perform request
	clientPointer := &http.Client{}
	ready.Lock()
	resp, err := clientPointer.Do(req)
	ready.Unlock()
	if err != nil {
		panic(err)
	}

	// Did not receive an OK or Accepted
	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		//		log.Printf("Request failed. Status code: %d\n", resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			panic(err)
		}

		//		log.Printf("Reply: %s\n", string(body))

		return body, resp.StatusCode, errors.New(string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		panic(err)
	}

	log.Printf("Reply: %#v\n", string(body))

	return body, 0, nil
}
