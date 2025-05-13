package search

import (
	"Tubes2_alchendol/models"
	"fmt"
	"sort"
	"time"
)

type TreeNode struct {
	Root     string     `json:"root"`
	Left     string     `json:"left"`
	Right    string     `json:"right"`
	Tier     string     `json:"tier"`
	Children []*TreeNode `json:"children"`
}

func processRecipes(recipes []models.Element, target string) (map[string][][]string, map[string]int) {
	foundTarget := false
	
	recipeMap := make(map[string][][]string)
	tierMap := make(map[string]int)

	// Untuk mendeteksi resep duplikat
	typeKeySet := make(map[string]map[string]bool)

	for _, r := range recipes {
		tierMap[r.Name] = r.Tier
		if len(r.Recipes) < 2 {
			continue
		}

		left := r.Recipes[0]
		right := r.Recipes[1]

		pair := []string{left, right}
		sort.Strings(pair)
		recipeKey := pair[0] + "+" + pair[1]

		if _, ok := typeKeySet[r.Name]; !ok {
			typeKeySet[r.Name] = make(map[string]bool)
		}

		if typeKeySet[r.Name][recipeKey] {
			continue
		}

		typeKeySet[r.Name][recipeKey] = true

		recipeMap[r.Name] = append(recipeMap[r.Name], []string{left, right})

		if r.Name == target {
			foundTarget = true
		}

		if foundTarget {
			if r.Name != target {
				break
			}		
		}
	}

	return recipeMap, tierMap
}

func BFS(target string, elements []models.Element) (*TreeNode, float64, int) {
	start := time.Now()

	recipeMap, tierMap := processRecipes(elements, target)
	
	recipes, exists := recipeMap[target]
	if !exists {
		return nil, time.Since(start).Seconds(), 0
	}
	

	// Iterasi elemen untuk mencari yang sesuai dengan target dan membangun pohon
	for _, recipe := range recipes {
		if len(recipe) != 2 {
			continue
		}

		left, right := recipe[0], recipe[1]
		leftTier, leftExists := tierMap[left]
		rightTier, rightExists := tierMap[right]

		if !leftExists || !rightExists || leftTier >= tierMap[target] || rightTier > tierMap[target] {
			continue
		}

		root := &TreeNode{
			Root:     target,
			Left:     left,
			Right:    right,
			Tier:     fmt.Sprintf("%d", tierMap[target]),
			Children: []*TreeNode{},
		}

		// Queue untuk BFS
		type queueItem struct {
			parentNode *TreeNode
			element    string
			tier       int
		}

		queue := []queueItem{
			{root, left, leftTier},
			{root, right, rightTier},
		}

		nodesVisited := 1 

		// Melakukan pencarian BFS
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			nodesVisited++

			// Jika elemen adalah dasar, tambahkan sebagai daun
			if IsBasicElement(current.element) {
				leaf := &TreeNode{
					Root:     current.element,
					Left:     "",
					Right:    "",
					Tier:     "0",
					Children: []*TreeNode{},
				}
				current.parentNode.Children = append(current.parentNode.Children, leaf)
				continue
			}

			// Mencari resep untuk elemen saat ini
			elementRecipes, exists := recipeMap[current.element]
			if !exists {
				continue
			}

			// Loop melalui resep untuk mencari yang bisa digunakan
			for _, r := range elementRecipes {
				if len(r) != 2 {
					continue
				}
				left, right := r[0], r[1]
				leftTier, leftOk := tierMap[left]
				rightTier, rightOk := tierMap[right]

				if !leftOk || !rightOk || leftTier >= current.tier || rightTier >= current.tier {
					continue
				}

				// Membuat node anak dari resep
				childNode := &TreeNode{
					Root:     current.element,
					Left:     left,
					Right:    right,
					Tier:     fmt.Sprintf("%d", current.tier),
					Children: []*TreeNode{},
				}
				current.parentNode.Children = append(current.parentNode.Children, childNode)

				// Menambahkan anak ke dalam queue
				queue = append(queue,
					queueItem{childNode, left, leftTier},
					queueItem{childNode, right, rightTier},
				)
				break // ambil satu resep saja
			}
		}

		return root, time.Since(start).Seconds(), nodesVisited
	}

	return nil, time.Since(start).Seconds(), 0
}
