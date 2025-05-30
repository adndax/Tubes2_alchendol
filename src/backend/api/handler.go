package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"Tubes2_alchendol/models"
	"Tubes2_alchendol/search"

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

	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target tidak boleh kosong"})
		return
	}

	elements, err := LoadRecipeData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal load recipes"})
		return
	}

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

    // Normalize algorithm name to lowercase for consistent matching
    algo = strings.ToLower(algo)
    
    switch algo {
    case "dfs":
        if multiple {
            fmt.Printf("[%s] Starting MultipleDFS with maxRecipes=%d\n", requestID, maxRecipes)
            
            // Multiple DFS in a separate goroutine
            go func() {
                defer func() {
                    if r := recover(); r != nil {
                        fmt.Printf("[%s] Recovered from panic: %v\n", requestID, r)
                        select {
                        case errChan <- fmt.Errorf("recovered from panic in MultipleDFS: %v", r):
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
                    "error": fmt.Sprintf("search timed out after %d seconds. the element '%s' (tier %d) is too complex or requires more time.", 
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
                        errChan <- fmt.Errorf("recovered from panic in DFS: %v", r)
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
                    "error": fmt.Sprintf("search timed out after %d seconds. the element '%s' (tier %d) is too complex or requires more time.", 
                        int(timeoutDuration.Seconds()), target, targetTier),
                    "target": target,
                    "requestId": requestID,
                    "isComplete": false,
                })
            }
        }
case "bfs":
    if multiple {
        fmt.Printf("[%s] Starting MultipleBFS with maxRecipes=%d\n", requestID, maxRecipes)
        
        // Multiple BFS in a separate goroutine
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    fmt.Printf("[%s] Recovered from panic: %v\n", requestID, r)
                    select {
                    case errChan <- fmt.Errorf("recovered from panic in MultipleBFS: %v", r):
                    case <-ctx.Done():
                        fmt.Printf("[%s] Context done while sending error\n", requestID)
                    }
                }
            }()
            
            fmt.Printf("[%s] Starting MultipleBFS\n", requestID)
            recipes, timeElapsed, nodesVisited := search.MultipleBFS(target, elements, maxRecipes)
            
            fmt.Printf("[%s] MultipleBFS completed with %d recipes\n", requestID, len(recipes))
            
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
                "error": fmt.Sprintf("search timed out after %d seconds. the element '%s' (tier %d) is too complex or requires more time.", 
                    int(timeoutDuration.Seconds()), target, targetTier),
                "target": target,
                "requestId": requestID,
                "isComplete": false,
            })
        }
        
        fmt.Printf("[%s] Request handling complete\n", requestID)
    } else {
        // Single BFS in a separate goroutine
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    errChan <- fmt.Errorf("recovered from panic in BFS: %v", r)
                }
            }()
            
            recipeTree, timeElapsed, nodesVisited := search.BFS(target, elements)
            
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
                "error": fmt.Sprintf("search timed out after %d seconds. the element '%s' (tier %d) is too complex or requires more time.", 
                    int(timeoutDuration.Seconds()), target, targetTier),
                "target": target,
                "requestId": requestID,
                "isComplete": false,
            })
        }
    }
    
	case "bidirectional":
		if multiple {
			// Multiple Bidirectional - now implemented
			fmt.Printf("[%s] Starting MultipleBidirectional with maxRecipes=%d\n", requestID, maxRecipes)
			
			// Multiple Bidirectional in a separate goroutine
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[%s] Recovered from panic: %v\n", requestID, r)
						select {
						case errChan <- fmt.Errorf("recovered from panic in MultipleBidirectional: %v", r):
						case <-ctx.Done():
							fmt.Printf("[%s] Context done while sending error\n", requestID)
						}
					}
				}()
				
				fmt.Printf("[%s] Starting MultipleBidirectional\n", requestID)
				recipes, timeElapsed, nodesVisited := search.MultipleBidirectional(target, elements, maxRecipes)
				
				fmt.Printf("[%s] MultipleBidirectional completed with %d recipes\n", requestID, len(recipes))
				
				// Format response the same way as MultipleDFS
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
			
			// Wait for result or timeout - same pattern as MultipleDFS
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
					"error": fmt.Sprintf("search timed out after %d seconds. the element '%s' (tier %d) is too complex or requires more time.", 
					int(timeoutDuration.Seconds()), target, targetTier),
					"target": target,
					"requestId": requestID,
					"isComplete": false,
				})
			}
			
			fmt.Printf("[%s] Request handling complete\n", requestID)
		} else {
			// Single Bidirectional in a separate goroutine
			go func() {
				defer func() {
					if r := recover(); r != nil {
						errChan <- fmt.Errorf("recovered from panic in BidirectionalSearch: %v", r)
					}
				}()
				
				recipeTree, timeElapsed, nodesVisited := search.BidirectionalSearch(target, elements)
				
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
                    "error": fmt.Sprintf("search timed out after %d seconds. the element '%s' (tier %d) is too complex or requires more time.", 
                    int(timeoutDuration.Seconds()), target, targetTier),
                    "target": target,
                    "requestId": requestID,
                    "isComplete": false,
                })
			}
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("algoritma tidak dikenali: %s", algo)})
	}
}