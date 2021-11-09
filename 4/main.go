package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)


const stringToSearch = "concurrency"

var sites = []string{
	"https://google.com",
	"https://itc.ua/",
	"https://twitter.com/concurrencyinc",
	"https://twitter.com/",
	"http://localhost:8000",
	"https://github.com/bradtraversy/go_restapi/blob/master/main.go",
	"https://www.youtube.com/",
	"https://postman-echo.com/get",
	"https://en.wikipedia.org/wiki/Concurrency_(computer_science)#:~:text=In%20computer%20science%2C%20concurrency%20is,without%20affecting%20the%20final%20outcome.",
}


type SiteData struct {
	data []byte
	uri  string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	//resultsCh := make(chan SiteData, len(sites))

	siteData := Worker(ctx, sites)
	go Reader(ctx, siteData, stringToSearch, cancel)
	//for _, uri := range sites {
	//	go parseSite(ctx, uri, stringToSearch, resultsCh)
	//}

	// give one second to validate if all other goroutines are closed
	time.Sleep(time.Second)
}

func Worker(ctx context.Context, sites []string) <-chan SiteData {
	resCh := make(chan SiteData, len(sites))
	for _, uri := range sites {
		go func(uri string) {
			res, err := request(ctx, uri)
			if err != nil {
				fmt.Println(err)
				return
			}

			resCh<-SiteData{data: res, uri: uri}
		}(uri)
	}

	return resCh
}


func Reader(ctx context.Context, dataCh <-chan SiteData, word string, cancel func()) {
	for {
		select {
		case data := <-dataCh:
			if findOccurance(data.data, word) {
				fmt.Println(fmt.Sprintf("'%v' string is found in: %v", word, data.uri))
				cancel()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// TODO implement function that will execute request function, will validate the output and cancel all other requests when needed page is found
// and will listen to cancellation signal from context and will exit from the func when will receive it
func parseSite(ctx context.Context, uri string, word string, result chan<- SiteData) {
	res, err := request(ctx, uri)
	if err != nil {
		fmt.Println(err)
		return
	}

	if findOccurance(res, word) {
		result <- SiteData{data: res, uri: uri}
		fmt.Println(fmt.Sprintf("'%v' string is found in: %v", word, uri))
		return
	}

	fmt.Println("Nothing found in: ", uri)
}


func findOccurance(data []byte, substr string) bool {
	return strings.Index(string(data), substr) != -1
}

func request(ctx context.Context, uri string) ([]byte, error){
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("Sending request to: %v", uri))
	resp, err := http.DefaultClient.Do(req)
	select{
	case <-ctx.Done():
		fmt.Println("Context canceled for: ", uri)
		return nil, nil
	default:
	}
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return bodyBytes, nil
}
