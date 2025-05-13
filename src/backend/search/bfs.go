package search

import (
	"Tubes2_alchendol/models"
	"fmt"
	"time"
	"sort"
)

// BFS performs a breadth-first search to find the shortest recipe for the target element
func BFS(target string, elements []models.Element) (models.RecipeTree, float64, int) {
	start := time.Now()

	recipeMap, tierMap := processRecipes(elements, target)
	
	// Jika target adalah elemen dasar, langsung buat recipe tree
	if IsBasicElement(target) {
		return createBasicElementLeaf(target), time.Since(start).Seconds(), 1
	}
	
	recipes, exists := recipeMap[target]
	if !exists {
		return models.RecipeTree{}, time.Since(start).Seconds(), 0
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

		root := createRecipeTreeNode(target, left, right, tierMap[target])

		// Queue untuk BFS
		type queueItem struct {
			parentNodeIdx int            // Index dari node parent dalam slice nodes
			element       string
			tier          int
		}

		// Slice untuk menyimpan semua node
		nodes := []models.RecipeTree{root}
		rootIdx := 0

		queue := []queueItem{
			{rootIdx, left, leftTier},
			{rootIdx, right, rightTier},
		}

		nodesVisited := 1 

		// Melakukan pencarian BFS
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			nodesVisited++

			// Jika elemen adalah dasar, tambahkan sebagai daun
			if IsBasicElementWithTier(current.element, tierMap) {
				leaf := createBasicElementLeaf(current.element)
				
				// Tambahkan leaf ke children dari parent node
				parentNode := &nodes[current.parentNodeIdx]
				parentNode.Children = append(parentNode.Children, leaf)
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
				childNode := createRecipeTreeNode(current.element, left, right, current.tier)
				
				// Tambahkan child node ke nodes slice
				nodes = append(nodes, childNode)
				childIdx := len(nodes) - 1
				
				// Tambahkan child node ke children dari parent node
				parentNode := &nodes[current.parentNodeIdx]
				parentNode.Children = append(parentNode.Children, childNode)

				// Menambahkan anak ke dalam queue
				queue = append(queue,
					queueItem{childIdx, left, leftTier},
					queueItem{childIdx, right, rightTier},
				)
				break // ambil satu resep saja
			}
		}

		// Deep copy node untuk menghindari referensi yang sama
		result := deepCopyRecipeTree(nodes[0])
		return result, time.Since(start).Seconds(), nodesVisited
	}

	return models.RecipeTree{}, time.Since(start).Seconds(), 0
}

// Fungsi helper yang dipakai oleh kedua algoritma (BFS dan MultipleBFS)
// Memisahkan ke utils.go agar tidak didefinisikan dua kali

// Fungsi untuk memproses resep dan membuat map untuk pencarian
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

// Deep copy untuk RecipeTree
func deepCopyRecipeTree(tree models.RecipeTree) models.RecipeTree {
	copied := models.RecipeTree{
		Root:     tree.Root,
		Left:     tree.Left,
		Right:    tree.Right,
		Tier:     tree.Tier,
		Children: []models.RecipeTree{},
	}
	
	for _, child := range tree.Children {
		copied.Children = append(copied.Children, deepCopyRecipeTree(child))
	}
	
	return copied
}

// Cek apakah elemen adalah elemen dasar
func IsBasicElementWithTier(element string, tierMap map[string]int) bool {
	if tier, exists := tierMap[element]; exists {
		return tier == 0
	}
	return IsBasicElement(element)
}

// Buat node daun (leaf node) untuk elemen dasar
func createBasicElementLeaf(element string) models.RecipeTree {
	return models.RecipeTree{
		Root:     element,
		Left:     "",
		Right:    "",
		Tier:     "0",
		Children: []models.RecipeTree{},
	}
}

// Membuat recipe tree node dengan data yang diberikan
func createRecipeTreeNode(element string, left string, right string, tier int) models.RecipeTree {
	return models.RecipeTree{
		Root:     element,
		Left:     left,
		Right:    right,
		Tier:     fmt.Sprintf("%d", tier),
		Children: []models.RecipeTree{},
	}
}