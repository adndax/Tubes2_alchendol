
/* Cara ngejalanin di WSL:
1.1. Pastikan Anda memiliki Go terinstal di sistem Anda.
	Cek versi Go dengan menjalankan perintah:
		> go version
	1.2. Jika Go belum terinstal, Anda dapat mengunduhnya dari https://golang.org/dl/ atau jalankan perintah berikut di terminal:
		> wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
		> sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz

	1.2.2 Go workspace (opsional), jalankan perintah berikut di terminal:
		> mkdir -p ~/go/{bin,src}

	1.3. Tambahkan Go ke PATH Anda dengan menambahkan baris berikut ke file ~/.bashrc atau ~/.bash_profile:
		> export GOPATH=$HOME/go
		> export GOBIN=$HOME/go/bin
		> export PATH=$PATH:/usr/local/go/bin:$GOBIN

	1.4. reload file dengan perintah:
		> source ~/.bashrc

	1.5. Cek apakah Go sudah terinstal dengan benar dengan menjalankan perintah:
		> go version
		> go env

2.1. Install Colly package dengan menjalankan perintah berikut di terminal:
	> go get github.com/gocolly/colly/...
	2.2. jika 'go get' tidak berhasil dan Anda menggunakan go versi 1.17 ke atas, coba jalankan perintah ini di direktori proyek Anda:
		> go mod init <nama_modul>
		lalu jalankan perintah 2.1.

3. Jalankan program di folder proyek dengan perintah:
   > go run .
*/

// package main

// import (
//     "encoding/json"
//     "fmt"
//     "os"
//     "strings"

//     "github.com/gocolly/colly"
// )

// type Element struct {
//     Name    string   `json:"name"`
//     Recipes []string `json:"recipes"`
// }

// func main() {
//     c := colly.NewCollector(
//         colly.AllowedDomains("little-alchemy.fandom.com"),
//     )
//     var elements []Element
//     for i := 2; i <= 18; i++ {
//         selector := fmt.Sprintf("div.mw-parser-output > table:nth-of-type(%d)", i)
//         c.OnHTML(selector, func(h *colly.HTMLElement) {
//             h.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
//                 element := row.ChildText("td:nth-of-type(1)")
//                 if element != "" {
//                     row.ForEach("td:nth-of-type(2) ul li", func(_ int, li *colly.HTMLElement) {
//                         recipe := strings.TrimSpace(li.Text)
//                         if recipe != "" {
//                             parts := strings.Split(recipe, " +  ")
//                             if len(parts) == 2 {
//                                 elements = append(elements, Element{
//                                     Name:    element,
//                                     Recipes: parts,
//                                 })
//                                 fmt.Printf("Processing element: %s, recipe: %v\n", element, parts)
//                             }
//                         }
//                     })
//                 }
//             })
//         })
//     }

//     c.Visit("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")

//     // Convert the elements slice to JSON
//     jsonData, err := json.MarshalIndent(elements, "", "  ")
//     if err != nil {
//         fmt.Println("Error marshalling JSON:", err)
//         return
//     }

//     // Write the JSON data to a .json file
//     file, err := os.Create("output.json")
//     if err != nil {
//         fmt.Println("Error creating file:", err)
//         return
//     }
//     defer file.Close()

//     _, err = file.WriteString(string(jsonData))
//     if err != nil {
//         fmt.Println("Error writing to file:", err)
//         return
//     }

//     fmt.Println("Scraped data has been written to output2.json")
// }