package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Replace with your directory path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}
	dirPath := filepath.Join(homeDir, "semester4/Tubes2_alchendol/src/images")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	listDir := make(map[string]int)
	for i, file := range files {
		name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		listDir[name] = i // Store the order of files
	}

	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
	res, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println("Error: Status code", res.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	var wg sync.WaitGroup
	doc.Find("table.list-table").Each(func(i int, table *goquery.Selection) {
		wg.Add(1)
		go func(table *goquery.Selection) {
			defer wg.Done()
			downloadSVG(table, listDir, dirPath)
		}(table)
	})

	wg.Wait()
}

func downloadSVG(table *goquery.Selection, listDir map[string]int, dirPath string) {
	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			// Skip the header row
			return
		}

		cols := row.Find("td")
		if cols.Length() < 2 {
			return
		}

		// Extract the name from the correct path
		name := cols.Eq(0).Find("a").Text()
		src, exists := cols.Eq(0).Find("a").Attr("href")
		if !exists {
			return
		}

		name = strings.TrimSpace(name)
		fmt.Println(name, listDir[name])

		if _, exists := listDir[name]; !exists {
			fmt.Println("[LOG] Downloading", name+".svg")
			resp, err := http.Get(src)
			if err != nil {
				fmt.Println("Error downloading image:", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Println("Error: Status code", resp.StatusCode)
				return
			}

			filePath := filepath.Join(dirPath, name+".svg")
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = io.Copy(file, resp.Body)
			if err != nil {
				fmt.Println("Error saving file:", err)
			}
		}
	})
}
