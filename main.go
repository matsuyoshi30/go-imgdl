package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	url := flag.String("url", "", "source page url")
	outdir := flag.String("o", "", "output dir")
	flag.Parse()

	client := &http.Client{}

	if *url == "" {
		fmt.Println("input url")
		return
	}
	req, err := http.NewRequest(http.MethodGet, *url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("status error")
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	bodyStr := string(body)
	mainStart := strings.Index(bodyStr, "<main")
	mainEnd := strings.Index(bodyStr, "</main")

	doc, err := html.Parse(strings.NewReader(bodyStr[mainStart-1 : mainEnd]))
	if err != nil {
		fmt.Println(err)
		return
	}
	imgPaths := extract(doc)

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}

	destDir := filepath.Join(currentDir, "dest")
	os.Mkdir(destDir, os.ModePerm)

	if *outdir != "" {
		destDir = filepath.Join(destDir, *outdir)
		err = os.Mkdir(destDir, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for idx, imgpath := range imgPaths {
		req, err := http.NewRequest(http.MethodGet, imgpath, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()

		file, err := os.Create(filepath.Join(destDir, fmt.Sprintf("%d.jpg", idx)))
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	fmt.Printf("Done! Downloaded %d\n", len(imgPaths))
}

var paths = make([]string, 0)

func extract(n *html.Node) []string {
	if n.Type == html.ElementNode && n.Data == "img" {
		for _, a := range n.Attr {
			if a.Key == "data-src" {
				paths = append(paths, a.Val)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extract(c)
	}

	return paths
}
