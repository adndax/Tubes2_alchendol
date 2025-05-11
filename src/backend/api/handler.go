package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

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
		}
	}

	println("Using multiple:", multiple)
	println("Max recipes:", maxRecipes)

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
