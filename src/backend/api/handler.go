package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
    "fmt"
    "time"

	"github.com/gin-gonic/gin"
	"Tubes2_alchendol/search"
    "Tubes2_alchendol/models"
)

func LoadRecipeData() ([]models.Element, error) {
	file, err := os.Open("data/elements.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var elements []models.Element
	err = json.NewDecoder(file).Decode(&elements)
	if err != nil {
		return nil, err
	}

	return elements, nil
}

func SearchHandler(c *gin.Context) {
    // Get parameters that match frontend expectations
    algo := c.Query("algo")
    target := strings.TrimSpace(c.Query("target"))
    mode := c.Query("mode") // "shortest" or "multiple"
    multipleStr := c.Query("multiple") // fallback parameter
    maxRecipesStr := c.Query("maxRecipes")

    // Debug logging
    fmt.Println("Received parameters:")
    fmt.Println("algo:", algo)
    fmt.Println("target:", target)
    fmt.Println("multiple:", multipleStr)
    fmt.Println("maxRecipes:", maxRecipesStr)

    if target == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Target tidak boleh kosong"})
        return
    }

    elements, err := LoadRecipeData()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal load recipes"})
        return
    }

    // Determine if multiple recipes requested
    // Check both mode parameter and multiple parameter
    multiple := false
    if mode == "multiple" {
        multiple = true
    } else if multipleStr == "true" {
        multiple = true
    }
    
    // If no mode or multiple parameter, check if maxRecipes > 1
    if mode == "" && multipleStr == "" && maxRecipesStr != "" {
        if max, err := strconv.Atoi(maxRecipesStr); err == nil && max > 1 {
            multiple = true
        }
    }

    // Get max recipes count
    maxRecipes := 5 // default
    if maxRecipesStr != "" {
        if max, err := strconv.Atoi(maxRecipesStr); err == nil {
            maxRecipes = max
            // Safety cap for maxRecipes
            if maxRecipes > 100 {
                maxRecipes = 100
            }
        }
    }

    fmt.Println("Using multiple:", multiple)
    fmt.Println("Max recipes:", maxRecipes)

    // Use a channel and timeout to handle potential slow or hanging requests
    resultChan := make(chan map[string]interface{}, 1)
    errChan := make(chan error, 1)
    
    // Start a timeout to prevent hanging requests
    timeout := time.After(10 * time.Second)

    switch algo {
    case "DFS":
        if multiple {
            // Multiple DFS in a separate goroutine
            go func() {
                defer func() {
                    // Recover from any panics to prevent the server from crashing
                    if r := recover(); r != nil {
                        errChan <- fmt.Errorf("Recovered from panic in MultipleDFS: %v", r)
                    }
                }()
                
                recipes, timeElapsed, nodesVisited := search.MultipleDFS(target, elements, maxRecipes)
                
                fmt.Printf("MultipleDFS returned %d recipes for %s (maxRecipes=%d)\n", 
                    len(recipes), target, maxRecipes)
                
                response := map[string]interface{}{
                    "nodesVisited": nodesVisited,
                    "roots": recipes,
                    "timeElapsed": timeElapsed,
                }
                
                resultChan <- response
            }()
            
            // Wait for result or timeout
            select {
            case result := <-resultChan:
                // Format and send the response
                c.Header("Content-Type", "application/json")
                prettyJSON, err := json.MarshalIndent(result, "", "    ")
                if err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
                    return
                }
                c.Writer.Write(prettyJSON)
                
            case err := <-errChan:
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                
            case <-timeout:
                c.JSON(http.StatusRequestTimeout, gin.H{
                    "error": "Waktu pencarian habis. Element terlalu kompleks atau search tree terlalu besar.",
                    "suggestion": "Coba element lain atau kurangi maxRecipes."})
            }
            
        } else {
            // Single DFS in a separate goroutine
            go func() {
                defer func() {
                    if r := recover(); r != nil {
                        errChan <- fmt.Errorf("Recovered from panic in DFS: %v", r)
                    }
                }()
                
                recipeTree, timeElapsed, nodesVisited := search.DFS(target, elements)
                
                response := map[string]interface{}{
                    "nodesVisited": nodesVisited,
                    "root": recipeTree,
                    "timeElapsed": timeElapsed,
                }
                
                resultChan <- response
            }()
            
            // Wait for result or timeout
            select {
            case result := <-resultChan:
                c.Header("Content-Type", "application/json")
                prettyJSON, err := json.MarshalIndent(result, "", "    ")
                if err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
                    return
                }
                c.Writer.Write(prettyJSON)
                
            case err := <-errChan:
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                
            case <-timeout:
                c.JSON(http.StatusRequestTimeout, gin.H{
                    "error": "Waktu pencarian habis. Element terlalu kompleks atau search tree terlalu besar."})
            }
        }
    
    case "BFS":
        // Add BFS implementation when available
        if multiple {
            // Multiple BFS - to be implemented
            c.JSON(http.StatusNotImplemented, gin.H{"error": "Multiple BFS belum diimplementasi"})
        } else {
            // Single BFS - to be implemented  
            c.JSON(http.StatusNotImplemented, gin.H{"error": "BFS belum diimplementasi"})
        }
    
    case "bidirectional", "Bidirectional", "BIDIRECTIONAL":
        if multiple {
            // Multiple Bidirectional - to be implemented
            c.JSON(http.StatusNotImplemented, gin.H{"error": "Multiple Bidirectional belum diimplementasi"})
        } else {
            // Single Bidirectional in a separate goroutine
            go func() {
                defer func() {
                    if r := recover(); r != nil {
                        errChan <- fmt.Errorf("Recovered from panic in BidirectionalSearch: %v", r)
                    }
                }()
                
                recipeTree, timeElapsed, nodesVisited := search.BidirectionalSearch(target, elements)
                
                response := map[string]interface{}{
                    "nodesVisited": nodesVisited,
                    "root": recipeTree,
                    "timeElapsed": timeElapsed,
                }
                
                resultChan <- response
            }()
            
            // Wait for result or timeout
            select {
            case result := <-resultChan:
                c.Header("Content-Type", "application/json")
                prettyJSON, err := json.MarshalIndent(result, "", "    ")
                if err != nil {
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
                    return
                }
                c.Writer.Write(prettyJSON)
                
            case err := <-errChan:
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                
            case <-timeout:
                c.JSON(http.StatusRequestTimeout, gin.H{
                    "error": "Waktu pencarian habis. Element terlalu kompleks atau search tree terlalu besar."})
            }
        }

    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "Algoritma tidak dikenali: " + algo})
    }
}