package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

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
		if val, err := strconv.Atoi(maxRecipesStr); err == nil && val > 0 && val <= 10 {
			maxRecipes = val
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maxRecipes harus berupa angka antara 1 sampai 10"})
			return
		}
	}

	switch algo {
	case "dfs":
		// Mode single hanya, DFS tidak mendukung multiple tree
		result := search.DFS(target, elements)
	
		c.Header("Content-Type", "application/json")
		prettyJSON, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
			return
		}
		c.Writer.Write(prettyJSON)
		
		case "bidirectional":
				if multiple {
					// Bidirectional Multiple - gunakan fungsi yang baru dibuat
					trees, timeTaken, nodes := search.MultipleBidirectional(target, elements, maxRecipes)
					
					// Format output untuk multiple recipes
					response := gin.H{
						"nodesVisited": nodes,
						"root": trees, // Array of trees
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