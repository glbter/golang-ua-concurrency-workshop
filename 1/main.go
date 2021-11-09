package main

import (
	"fmt"
	"github.com/glbter/golang-ua-concurency-workshop/1/stream"
	"time"
)

func producer(str stream.Stream) <-chan stream.Tweet {
	tweets := make(chan stream.Tweet)
	go func(str stream.Stream) {
		for {
			tweet, err := str.Next()
			if err == stream.ErrEOF {
				close(tweets)
				return
			}

			tweets <- *tweet
		}
	}(str)

	return tweets
}

func producer2(str stream.Stream, tweets chan<- stream.Tweet) {
	for {
		tweet, err := str.Next()
		if err == stream.ErrEOF {
			close(tweets)
			return
		}

		tweets <- *tweet
	}
}

func consumer(tweets <-chan stream.Tweet) {
	for t := range tweets {
		if t.IsTalkingAboutGo() {
			fmt.Println(t.Username, "\ttweets about golang")
			continue
		}

		fmt.Println(t.Username, "\tdoes not tweet about golang")
	}
}

func consumer2(tweets <-chan stream.Tweet, done chan<- bool) {
	for t := range tweets {
		if t.IsTalkingAboutGo() {
			fmt.Println(t.Username, "\ttweets about golang")
			continue
		}

		fmt.Println(t.Username, "\tdoes not tweet about golang")
	}

	done <- true
}

func main() {
	start := time.Now()
	str := stream.GetMockStream()

	tweets := producer(str)

	consumer(tweets)

	fmt.Printf("Process took %s\n", time.Since(start))

	fmt.Println()
//// variant2
//
//	start = time.Now()
//	str = stream.GetMockStream()
//
//	tweets2 := make(chan stream.Tweet)
//	done := make(chan bool)
//
//	go consumer2(tweets2, done)
//	go producer2(str, tweets2)
//	<-done
//
//	fmt.Printf("Process took %s\n", time.Since(start))
}
