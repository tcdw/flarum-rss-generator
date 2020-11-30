package main

import (
	"encoding/json"
	"fmt"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/buger/jsonparser"
)

func fatalError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	var opts struct {
		Site string `short:"s" long:"site" description:"The Flarum site you want to get RSS from" required:"true"`
		Output string `short:"o" long:"output" description:"Path of the output file" required:"true"`
	}

	_, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	// 下载原始数据
	fullURL := fmt.Sprintf("%s/api/discussions?include=user%%2ClastPostedUser%%2Ctags%%2CfirstPost&sort=-createdAt&page%%5Boffset%%5D", opts.Site)
	req, err := http.NewRequest("GET", fullURL, nil)
	fatalError(err)

	req.Header.Set("User-Agent", "Flarum_RSS_Bot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	fatalError(err)

	body, err := ioutil.ReadAll(resp.Body)
	fatalError(err)

	// WIP
}
