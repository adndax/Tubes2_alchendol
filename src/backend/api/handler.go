package api

import (
	"encoding/json"
	"net/http"
	"os"
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
	algo := c.Query("algo")
	target := strings.TrimSpace(c.Query("target"))

	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target tidak boleh kosong"})
		return
	}

	elements, err := LoadRecipeData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal load recipes"})
		return
	}

	switch algo {
	case "DFS":
		result := search.DFS(target, elements)

		// Manual JSON output
		c.Header("Content-Type", "application/json")
		prettyJSON, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memformat JSON"})
			return
		}
		c.Writer.Write(prettyJSON)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Algoritma tidak dikenali"})
	}
}