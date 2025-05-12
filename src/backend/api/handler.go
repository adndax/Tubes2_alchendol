package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
<<<<<<< HEAD

	"Tubes2_alchendol/models"
	"Tubes2_alchendol/search"
=======
    "fmt"
    "time"
    "context"
>>>>>>> 94822aee96831c6a3f0a6783653fa86225694b2a

	"github.com/gin-gonic/gin"
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
<<<<<<< HEAD
	// Get parameters that match frontend expectations
	algo := c.Query("algo")
	target := strings.TrimSpace(c.Query("target"))
	mode := c.Query("mode")            // "shortest" or "multiple"
	multipleStr := c.Query("multiple") // fallback parameter
	maxRecipesStr := c.Query("maxRecipes")

	// Debug logging
	println("Received parameters:")
	println("algo:", algo)
	println("target:", target)
	println("multiple:", multipleStr)
	println("maxRecipes:", maxRecipesStr)
=======
    // Add request ID for tracking
    requestID := fmt.Sprintf("%d", time.Now().UnixNano())
    fmt.Printf("[%s] Starting request processing\n", requestID)

    // Get parameters that match frontend expectations
    algo := c.Query("algo")
    target := strings.TrimSpace(c.Query("target"))
    mode := c.Query("mode") // "shortest" or "multiple"
    multipleStr := c.Query("multiple") // fallback parameter
    maxRecipesStr := c.Query("maxRecipes")

    // Debug logging
    fmt.Printf("[%s] Received parameters: algo=%s, target=%s, mode=%s, multiple=%s, maxRecipes=%s\n", 
        requestID, algo, target, mode, multipleStr, maxRecipesStr)
>>>>>>> 94822aee96831c6a3f0a6783653fa86225694b2a

	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target tidak boleh kosong"})
		return
	}

	elements, err := LoadRecipeData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal load recipes"})
		return
	}

<<<<<<< HEAD
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
		}
	}

	println("Using multiple:", multiple)
	println("Max recipes:", maxRecipes)
=======
    // Determine target element's tier for timeout calculation
    targetTier := 0
    for _, el := range elements {
        if el.Name == target {
            targetTier = el.Tier
            break
        }
    }

    // Determine if multiple recipes requested
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

    fmt.Printf("[%s] Using multiple=%v, maxRecipes=%d, targetTier=%d\n", 
        requestID, multiple, maxRecipes, targetTier)

    // Determine timeout based on tier
    // Higher tier elements get more time
    timeoutDuration := 30 * time.Second
    if targetTier > 5 {
        timeoutDuration = 60 * time.Second
    }
    
    // Special case for known complex elements
    complexElements := map[string]bool{
        "Picnic": true, "Skyscraper": true, "City": true, 
        "Continent": true, "Horseshoe": true, "Unicorn": true,
        "Human": true, "Astronaut": true, "Doctor": true,
    }
    
    if complexElements[target] {
        timeoutDuration = 60 * time.Second
    }
    
    fmt.Printf("[%s] Setting timeout to %v for element tier %d\n", 
        requestID, timeoutDuration, targetTier)

    // Create parent context with timeout
    ctx, cancel := context.WithTimeout(c.Request.Context(), timeoutDuration)
    defer cancel()

    // Use a channel for results
    resultChan := make(chan map[string]interface{}, 1)
    errChan := make(chan error, 1)

    switch algo {
    case "DFS":
        if multiple {
            fmt.Printf("[%s] Starting MultipleDFS with maxRecipes=%d\n", requestID, maxRecipes)
            
            // Multiple DFS in a separate goroutine
            go func() {
                defer func() {
                    if r := recover(); r != nil {
                        fmt.Printf("[%s] Recovered from panic: %v\n", requestID, r)
                        select {
                        case errChan <- fmt.Errorf("Recovered from panic in MultipleDFS: %v", r):
                        case <-ctx.Done():
                            fmt.Printf("[%s] Context done while sending error\n", requestID)
                        }
                    }
                }()
                
                fmt.Printf("[%s] Starting MultipleDFS\n", requestID)
                recipes, timeElapsed, nodesVisited := search.MultipleDFS(target, elements, maxRecipes)
                
                fmt.Printf("[%s] MultipleDFS completed with %d recipes\n", requestID, len(recipes))
                
                // Always send a response, even if no recipes found or fewer than requested
                response := map[string]interface{}{
                    "nodesVisited": nodesVisited,
                    "roots": recipes,
                    "timeElapsed": timeElapsed,
                    "requestId": requestID,
                    "recipesCount": len(recipes),
                    "maxRequested": maxRecipes,
                    "target": target,
                    "targetTier": targetTier,
                    "isComplete": true,
                }
                
                select {
                case resultChan <- response:
                    fmt.Printf("[%s] Sent response to channel\n", requestID)
                case <-ctx.Done():
                    fmt.Printf("[%s] Context done while sending result\n", requestID)
                }
            }()
            
            // Wait for result or timeout
            select {
            case result := <-resultChan:
                fmt.Printf("[%s] Received result, sending to client\n", requestID)
                c.Header("Content-Type", "application/json")
                c.Header("Cache-Control", "no-store, no-cache")
                c.Header("X-Request-ID", requestID)
                c.JSON(http.StatusOK, result)
                
            case err := <-errChan:
                fmt.Printf("[%s] Received error: %v\n", requestID, err)
                c.Header("Cache-Control", "no-store, no-cache")
                c.Header("X-Request-ID", requestID)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": err.Error(),
                    "requestId": requestID,
                    "isComplete": true,
                })
                
            case <-ctx.Done():
                fmt.Printf("[%s] Request timed out\n", requestID)
                c.Header("Cache-Control", "no-store, no-cache")
                c.Header("X-Request-ID", requestID)
                c.JSON(http.StatusRequestTimeout, gin.H{
                    "error": fmt.Sprintf("Search timed out after %d seconds. The element '%s' (tier %d) is too complex or requires more time.", 
                        int(timeoutDuration.Seconds()), target, targetTier),
                    "target": target,
                    "requestId": requestID,
                    "isComplete": false,
                })
            }
            
            fmt.Printf("[%s] Request handling complete\n", requestID)
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
                    "requestId": requestID,
                    "target": target,
                    "targetTier": targetTier,
                    "isComplete": true,
                }
                
                resultChan <- response
            }()
            
            // Wait for result or timeout
            select {
            case result := <-resultChan:
                c.Header("Content-Type", "application/json")
                c.Header("Cache-Control", "no-store, no-cache")
                c.Header("X-Request-ID", requestID)
                c.JSON(http.StatusOK, result)
                
            case err := <-errChan:
                c.Header("Cache-Control", "no-store, no-cache")
                c.Header("X-Request-ID", requestID)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": err.Error(),
                    "requestId": requestID,
                    "isComplete": false,
                })
                
            case <-ctx.Done():
                c.Header("Cache-Control", "no-store, no-cache")
                c.Header("X-Request-ID", requestID)
                c.JSON(http.StatusRequestTimeout, gin.H{
                    "error": fmt.Sprintf("Search timed out after %d seconds. The element '%s' (tier %d) is too complex or requires more time.", 
                        int(timeoutDuration.Seconds()), target, targetTier),
                    "target": target,
                    "requestId": requestID,
                    "isComplete": false,
                })
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
                
            case <-ctx.Done():
                c.JSON(http.StatusRequestTimeout, gin.H{
                    "error": "Waktu pencarian habis. Element terlalu kompleks atau search tree terlalu besar."})
            }
        }
>>>>>>> 94822aee96831c6a3f0a6783653fa86225694b2a

	switch algo {
	case "DFS":
		if multiple {
			// Multiple DFS
			recipes, timeElapsed, nodesVisited := search.MultipleDFS(target, elements, maxRecipes)

			// Debug: Check what we got back
			fmt.Printf("DEBUG: MultipleDFS returned %d recipes for %s (maxRecipes=%d)\n",
				len(recipes), target, maxRecipes)

			appendRecipes := map[string]interface{}{
				"nodesVisited": nodesVisited,
				"roots":        recipes,
				"timeElapsed":  timeElapsed,
			}

			// Use same format as single DFS
			response := gin.H{
				"nodesVisited": appendRecipes["nodesVisited"],
				"roots":        appendRecipes["roots"],
				"timeElapsed":  appendRecipes["timeElapsed"],
			}

			c.Header("Content-Type", "application/json")
			prettyJSON, err := json.MarshalIndent(response, "", "    ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
				return
			}
			c.Writer.Write(prettyJSON)
		} else {
			// Single DFS (remains the same)
			recipeTree, timeElapsed, nodesVisited := search.DFS(target, elements)

			response := gin.H{
				"nodesVisited": nodesVisited,
				"root":         recipeTree,
				"timeElapsed":  timeElapsed,
			}

			c.Header("Content-Type", "application/json")
			prettyJSON, err := json.MarshalIndent(response, "", "    ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
				return
			}
			c.Writer.Write(prettyJSON)
		}

	case "BFS":
		if multiple {
			// Multiple BFS
			recipes, timeElapsed, nodesVisited := search.MultipleBFS(target, elements, maxRecipes)

			// Debug: Check what we got back
			fmt.Printf("DEBUG: MultipleDFS returned %d recipes for %s (maxRecipes=%d)\n",
				len(recipes), target, maxRecipes)

			appendRecipes := map[string]interface{}{
				"nodesVisited": nodesVisited,
				"roots":        recipes,
				"timeElapsed":  timeElapsed,
			}

			// Use same format as single BFS
			response := gin.H{
				"nodesVisited": appendRecipes["nodesVisited"],
				"roots":        appendRecipes["roots"],
				"timeElapsed":  appendRecipes["timeElapsed"],
			}

			c.Header("Content-Type", "application/json")
			prettyJSON, err := json.MarshalIndent(response, "", "    ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
				return
			}
			c.Writer.Write(prettyJSON)
		} else {
			recipeTree, timeElapsed, nodesVisited := search.BFS(target, elements)

			response := gin.H{
				"nodesVisited": nodesVisited,
				"root":         recipeTree,
				"timeElapsed":  timeElapsed,
			}

			c.Header("Content-Type", "application/json")
			prettyJSON, err := json.MarshalIndent(response, "", "    ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
				return
			}
			c.Writer.Write(prettyJSON)
		}

	case "bidirectional", "Bidirectional":
		if multiple {
			// Multiple Bidirectional - to be implemented
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Multiple Bidirectional belum diimplementasi"})
		} else {
			// Single Bidirectional
			recipeTree, timeElapsed, nodesVisited := search.BidirectionalSearch(target, elements)

			// Format response with single root
			response := gin.H{
				"nodesVisited": nodesVisited,
				"root":         recipeTree,
				"timeElapsed":  timeElapsed,
			}

			c.Header("Content-Type", "application/json")
			prettyJSON, err := json.MarshalIndent(response, "", "    ")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
				return
			}
			c.Writer.Write(prettyJSON)
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Algoritma tidak dikenali: " + algo})
	}
}
