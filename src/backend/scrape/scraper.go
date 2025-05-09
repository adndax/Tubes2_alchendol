package scrape

import (
    "encoding/json"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"

    "github.com/gocolly/colly"
    "Tubes2_alchendol/models"
)

// ScrapeElements scrapes the elements and writes them to a JSON file
func ScrapeElements(outputFile string) error {
    c := colly.NewCollector(
        colly.AllowedDomains("little-alchemy.fandom.com"),
    )

    // Map to store the tier of each element
    elementTiers := make(map[string]int)
    var elements []models.Element

    // Collect all elements and their tiers
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
                    }
                })
            }
        })
    })
    c.Visit("https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)")

    // Add basic elements (tier 0) manually with empty recipes
    basicElements := []string{"Air", "Earth", "Fire", "Water"}
    for _, name := range basicElements {
        elements = append(elements, models.Element{
            Name:    name,
            Recipes: []string{}, // Empty recipes for basic elements
            Tier:    0,
        })
        elementTiers[name] = 0
        fmt.Printf("Added basic element: %s with tier 0\n", name)
    }

    // Collect all recipes for non-basic elements
    d := colly.NewCollector(
        colly.AllowedDomains("little-alchemy.fandom.com"),
    )

    d.OnHTML("div.mw-parser-output > table", func(h *colly.HTMLElement) {
        h.ForEach("tbody tr", func(_ int, row *colly.HTMLElement) {
            element := strings.TrimSpace(row.ChildText("td:nth-of-type(1)"))
            if element != "" && !contains(basicElements, element) { // Skip basic elements
                row.ForEach("td:nth-of-type(2) ul li", func(_ int, li *colly.HTMLElement) {
                    recipe := strings.TrimSpace(li.Text)
                    if recipe != "" {
                        var parts []string
                        parts = strings.Split(recipe, " + ")

                        if len(parts) == 2 {
                            for i := range parts {
                                parts[i] = strings.TrimSpace(parts[i])
                            }

                            tier := elementTiers[element]

                            elements = append(elements, models.Element{
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
        return fmt.Errorf("error marshalling JSON: %w", err)
    }

    // Write the JSON data to a .json file
    file, err := os.Create(outputFile)
    if err != nil {
        return fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    _, err = file.WriteString(string(jsonData))
    if err != nil {
        return fmt.Errorf("error writing to file: %w", err)
    }

    fmt.Printf("Scraped data has been written to %s\n", outputFile)
    fmt.Printf("Total elements processed: %d\n", len(elements))
    fmt.Printf("Total elements with tiers: %d\n", len(elementTiers))

    return nil
}

// Helper function to check if a string is in a slice
func contains(slice []string, str string) bool {
    for _, v := range slice {
        if v == str {
            return true
        }
    }
    return false
}