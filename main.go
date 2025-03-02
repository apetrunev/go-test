// comment
package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

func main() {
	url := "http://kinopoisk.ru"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	buf := bytes.NewBuffer(b)
	z := html.NewTokenizer(buf)
	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			return
		}
		fmt.Printf("%v+\n", tt)
	}

}
