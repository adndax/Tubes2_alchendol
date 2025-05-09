package main

import (
    "encoding/json"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"

    "github.com/gocolly/colly"
)

type Element struct {
    Name    string   `json:"name"`
    Recipes []string `json:"recipes"`
    Tier    int      `json:"tier"`
}

func main() {
    c := colly.NewCollector(
        colly.AllowedDomains("little-alchemy.fandom.com"),
    )
    
    // Map to store the tier of each element
    elementTiers := make(map[string]int)
    var elements []Element
    
    // collect all elements and their tiers
    c.OnHTML("div.mw-parser-output", func(h *colly.HTMLElement) {
        currentTier := 0
        tierRegex := regexp.MustCompile(`Tier (\d+)`)
        
        h.ForEach("h3, table", func(_ int, element *colly.HTMLElement) {
            if element.Name == "h3" {
                headerText := element.Text
                matches := tierRegex.FindStringSubmatch(headerText)
                if len(matches) > 1 {
                    tier, err := strconv.Atoi(matches[1])
                    if err == nil {
                        currentTier = tier
                        fmt.Printf("Found Tier %d\n", currentTier)
                    }
                }
            } else if element.Name == "table" && currentTier > 0 {
                // This is a table under a tier header
                element.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
                    elementName := strings.TrimSpace(row.ChildText("td:nth-of-type(1)"))
                    if elementName != "" {
                        elementTiers[elementName] = currentTier
                        //fmt.Printf("Element %s is in Tier %d\n", elementName, currentTier)
                    }
                })
            }
        })
    })
    c.Visit("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")
    
    // collect all recipes and associate with tiers
    d := colly.NewCollector(
        colly.AllowedDomains("little-alchemy.fandom.com"),
    )
    
    d.OnHTML("div.mw-parser-output > table", func(h *colly.HTMLElement) {
        h.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
            element := strings.TrimSpace(row.ChildText("td:nth-of-type(1)"))
            if element != "" {
                row.ForEach("td:nth-of-type(2) ul li", func(_ int, li *colly.HTMLElement) {
                    recipe := strings.TrimSpace(li.Text)
                    if recipe != "" {
                        // The recipe format might vary, try different separators
                        var parts []string
                        if strings.Contains(recipe, " + ") {
                            parts = strings.Split(recipe, " + ")
                        } else if strings.Contains(recipe, " +  ") {
                            parts = strings.Split(recipe, " +  ")
                        } else {
                            // Try to find a + surrounded by spaces using regex
                            re := regexp.MustCompile(`\s+\+\s+`)
                            parts = re.Split(recipe, -1)
                        }
                        
                        if len(parts) == 2 {
                            // Clean up the parts and make sure they are properly trimmed
                            for i := range parts {
                                parts[i] = strings.TrimSpace(parts[i])
                            }
                            
                            // Get the tier for this element
                            tier := elementTiers[element]
                            
                            elements = append(elements, Element{
                                Name:    element,
                                Recipes: parts,
                                Tier:    tier,
                            })
                            fmt.Printf("Processing element: %s, recipe: %v, tier: %d\n", element, parts, tier)
                        }
                    }
                })
            }
        })
    })
    
    d.Visit("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")
    
    // Convert the elements slice to JSON
    jsonData, err := json.MarshalIndent(elements, "", "  ")
    if err != nil {
        fmt.Println("Error marshalling JSON:", err)
        return
    }
    
    // Write the JSON data to a .json file
    file, err := os.Create("../data/elements.json")
    if err != nil {
        fmt.Println("Error creating file:", err)
        return
    }
    defer file.Close()
    
    _, err = file.WriteString(string(jsonData))
    if err != nil {
        fmt.Println("Error writing to file:", err)
        return
    }
    
    fmt.Println("Scraped data has been written to elements.json")
    fmt.Printf("Total elements processed: %d\n", len(elements))
    fmt.Printf("Total elements with tiers: %d\n", len(elementTiers))
}