package main

import (
	// "github.com/gin-contrib/cors" 
	// "github.com/gin-gonic/gin"
	"Tubes2_alchendol/api"
	"Tubes2_alchendol/search"
	// "Tubes2_alchendol/scrape"
	"fmt"
)

func main() {
	// Contoh penggunaan fungsi DFS
	// Membaca elemen dari file JSON
	// err := scrape.ScrapeElements("data/elements.json")

	// if err != nil {
	// 	fmt.Printf("Error scraping elements: %v\n", err)
	// 	return
	// }

	elements, err := api.LoadRecipeData()
	if err != nil {
		fmt.Printf("Error membaca elemen: %v\n", err)
		return
	}

    // elementMap, targetFound := search.CreateFilteredElementMap(elements, "Dust")  // Perlu mengekspos fungsi ini sebagai publik
    // if targetFound {
	// 	fmt.Printf("\nElementMap contains: %+v\n", elementMap)
	// 	return
	// }
	

	result := search.DFS("Firetruck", elements)
		fmt.Printf("Hasil pencarian: %+v\n", result)
	

}

