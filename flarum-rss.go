package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/buger/jsonparser"
	"github.com/gorilla/feeds"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func fatalError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

func getMeta(url string) (title string, description string, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", "Flarum_RSS_Bot/1.0")
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	siteTitle := doc.Find("head > title").Text()
	siteDescription := ""
	siteDescriptionElement := doc.Find("head > meta[name=description]")
	siteDescriptionData, exists := siteDescriptionElement.Attr("content")
	if exists {
		siteDescription = siteDescriptionData
	}
	return siteTitle, siteDescription, nil
}

func getThreads(url string) (data []byte, err error) {
	fullURL := fmt.Sprintf("%s/api/discussions?include=user%%2ClastPostedUser%%2Ctags%%2CfirstPost&sort=-createdAt&page%%5Boffset%%5D", url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Flarum_RSS_Bot/1.0")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func main() {
	var opts struct {
		Site string `short:"s" long:"site" description:"The Flarum site you want to get RSS from" required:"true"`
		Output string `short:"o" long:"output" description:"Path of the output file" required:"true"`
		Type string `short:"t" long:"type" description:"Feed format" choice:"rss" choice:"feedData"`
	}

	_, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	log.Println("Retrieving site meta data")
	site := opts.Site
	title, description, err := getMeta(site)
	fatalError(err)
	log.Printf("Title: %s\n", title)
	log.Printf("Description: %s\n", description)

	log.Println("Retrieving thread list")
	data, err := getThreads(site)
	fatalError(err)

	now := time.Now()
	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: site},
		Description: description,
		Created:     now,
	}
	feed.Items = []*feeds.Item{}

	_, err = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		// 获取基本信息
		postID, _ := jsonparser.GetString(value, "id")
		title, _ := jsonparser.GetString(value, "attributes", "title")
		createdAt, _ := jsonparser.GetString(value, "attributes", "createdAt")
		userID, _ := jsonparser.GetString(value, "relationships", "user", "data", "id")
		firstPost, _ := jsonparser.GetString(value, "relationships", "firstPost", "data", "id")
		// 获取主题作者信息
		var authorName string
		var content string
		_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			cType, _ := jsonparser.GetString(value, "type")
			switch cType {
			case "users":
				cUserID, _ := jsonparser.GetString(value, "id")
				if cUserID == userID {
					authorName, _ = jsonparser.GetString(value, "attributes", "displayName")
				}
				break
			case "posts":
				cFirstPost, _ := jsonparser.GetString(value, "id")
				if cFirstPost == firstPost {
					content, _ = jsonparser.GetString(value, "attributes", "contentHtml")
				}
				break
			}
		}, "included")
		created, _ := time.Parse(time.RFC3339, createdAt);
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/d/%s", site, postID)},
			Description: content,
			Author:      &feeds.Author{Name: authorName},
			Created:     created,
		})
	}, "data")
	fatalError(err)

	var feedData string

	switch opts.Type {
	case "rss":
		feedData, err = feed.ToAtom()
		fatalError(err)
		break
	case "atom":
	default:
		feedData, err = feed.ToRss()
		fatalError(err)
		break
	}

	if opts.Output == "-" {
		_, err = os.Stdout.Write([]byte(feedData))
		fatalError(err)
		return
	}
	err = ioutil.WriteFile(opts.Output, []byte(feedData), 0644)
	fatalError(err)
}
