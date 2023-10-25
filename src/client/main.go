package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func httpPost(client *http.Client, url string, data string) {
	tm1 := time.Now().UnixMicro()

	reqBody := strings.NewReader(data)
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	//var client = http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	BodyData, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(BodyData))

	tm2 := time.Now().UnixMicro()
	delta := tm2 - tm1
	fmt.Printf("cost: %d ms, url = %s\n\n", delta/1000.0, url)
}

func testGpx(client *http.Client) {
	url := "http://localhost:7817/v1/gpx"
	data1 := `{"phone":"13800138000","lat":40.12,"lon":116.123,"ele":10,"tm":`
	data2 := `,"speed":0}`
	tm := time.Now().Unix()
	tmStr := strconv.FormatInt(tm, 10)
	data := data1 + tmStr + data2

	httpPost(client, url, data)

}

func testGetLast(client *http.Client) {
	url := "http://localhost:7817/v1/position"
	data := `{"phone":"13800138000","friend":"13800138000"}`
	httpPost(client, url, data)
}

func testGetTrack(client *http.Client) {
	url := `http://localhost:7817/v1/track`
	tm := time.Now().Unix() - 30
	tmStr := fmt.Sprintf("%d", tm)

	data := `{"phone":"13800138000","friend":"13800138000", "date":"20220923", "tmStart":` + tmStr + `}`
	data = `{"phone":"13800138000","friend":"13800138000", "date":"20220923" }`
	fmt.Println(data)
	httpPost(client, url, data)
}

func main() {
	httpClient := &http.Client{}

	for i := 0; i < 100; i++ {

		testGpx(httpClient)
		testGetLast(httpClient)
		testGetTrack(httpClient)
		time.Sleep(5 * time.Second)
	}
}
