package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"fmt"

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
	algo := strings.ToLower(c.Query("algo"))
	target := strings.TrimSpace(c.Query("target"))
	multiple := c.Query("multiple") == "true"
	maxRecipesStr := c.DefaultQuery("maxRecipes", "3")

	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target tidak boleh kosong"})
		return
	}

	elements, err := LoadRecipeData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal load recipes"})
		return
	}

	// Cek jumlah maksimum resep
	maxRecipes := 3
	if multiple {
		if val, err := strconv.Atoi(maxRecipesStr); err == nil {
			maxRecipes = val
		}
	}

	switch algo {
	case "dfs":
        if multiple {
            // Multiple DFS
            recipes, timeElapsed, nodesVisited := search.MultipleDFS(target, elements, maxRecipes)
            
            // Debug: Check what we got back
            fmt.Printf("DEBUG: MultipleDFS returned %d recipes for %s (maxRecipes=%d)\n", 
                len(recipes), target, maxRecipes)
                
            // Use consistent format for both algorithms
            response := gin.H{
                "nodesVisited": nodesVisited,
                "roots": recipes,
                "timeElapsed": timeElapsed,
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
                "root": recipeTree,
                "timeElapsed": timeElapsed,
            }
            
            c.Header("Content-Type", "application/json")
            prettyJSON, err := json.MarshalIndent(response, "", "    ")
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
                return
            }
            c.Writer.Write(prettyJSON)
        }
    
		
    case "bidirectional":
        if multiple {
            // Bidirectional Multiple - gunakan fungsi yang baru dibuat
            trees, timeTaken, nodes := search.MultipleBidirectional(target, elements, maxRecipes)
            
            // Format output untuk multiple recipes - use the same format as DFS for consistency
            response := gin.H{
                "nodesVisited": nodes,
                "roots": trees, // Use 'roots' key instead of 'root' for consistency with DFS
                "timeElapsed": timeTaken,
            }
            
            c.Header("Content-Type", "application/json")
            prettyJSON, err := json.MarshalIndent(response, "", "    ")
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
                return
            }
            c.Writer.Write(prettyJSON)
        } else {
            // Bidirectional Single
            tree, timeTaken, nodes := search.BidirectionalSearch(target, elements)
            
            response := gin.H{
                "nodesVisited": nodes,
                "root": tree,
                "timeElapsed": timeTaken,
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
        c.JSON(http.StatusBadRequest, gin.H{"error": "Algoritma tidak dikenali. Gunakan 'dfs' atau 'bidirectional'"})
    }
}