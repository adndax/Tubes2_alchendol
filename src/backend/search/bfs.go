package search

import (
	"Tubes2_alchendol/models"
	"container/list"
	"fmt"
	"time"
	"sort"
)

// BFS performs a breadth-first search to find the shortest recipe for the target element
func BFS(target string, elements []models.Element) (models.RecipeTree, float64, int) {
	start := time.Now()

	// Filter elements dan membuat map
	recipeMap, tierMap := processRecipes(elements, target)
	
	// Jika target adalah elemen dasar, langsung buat recipe tree
	if IsBasicElement(target) {
		return createBasicElementLeaf(target), time.Since(start).Seconds(), 1
	}
	
	// Cek apakah target ada di resep map
	recipes, exists := recipeMap[target]
	if !exists {
		return models.RecipeTree{}, time.Since(start).Seconds(), 0
	}

	// Cari salah satu resep valid untuk target
	for _, recipe := range recipes {
		if len(recipe) != 2 {
			continue
		}

		left, right := recipe[0], recipe[1]
		leftTier, leftExists := tierMap[left]
		rightTier, rightExists := tierMap[right]

		// Skip jika kombinasi tidak valid
		if !leftExists || !rightExists || leftTier >= tierMap[target] || rightTier > tierMap[target] {
			continue
		}

		// Buat root node
		root := models.RecipeTree{
			Root:     target,
			Left:     left,
			Right:    right,
			Tier:     fmt.Sprintf("%d", tierMap[target]),
			Children: []models.RecipeTree{},
		}

		// Gunakan BFS yang benar dengan container/list untuk queue
		nodesVisited := 1 // Mulai dengan root node
		
		// Lakukan BFS untuk setiap komponen
		comp1Tree := bfsSearchComponentToBasics(left, recipeMap, tierMap, &nodesVisited)
		comp2Tree := bfsSearchComponentToBasics(right, recipeMap, tierMap, &nodesVisited)
		
		// Jika salah satu component tidak berhasil dibuat, lanjut ke resep berikutnya
		if comp1Tree.Root == "" || comp2Tree.Root == "" {
			continue
		}
		
		// Tambahkan component trees sebagai children
		root.Children = append(root.Children, comp1Tree, comp2Tree)
		
		// Debug: Melihat struktur resep
		// fmt.Printf("Recipe for %s: %s + %s\n", target, left, right)
		
		return root, time.Since(start).Seconds(), nodesVisited
	}

	// Jika tidak ada resep valid, kembalikan tree kosong
	return models.RecipeTree{}, time.Since(start).Seconds(), 0
}

// BFS untuk mencari path dari component ke elemen dasar
func bfsSearchComponentToBasics(element string, recipeMap map[string][][]string, tierMap map[string]int, nodesVisited *int) models.RecipeTree {
	// Increment counter
	(*nodesVisited)++
	
	// Jika elemen adalah dasar, langsung kembalikan
	if IsBasicElementWithTier(element, tierMap) {
		return createBasicElementLeaf(element)
	}
	
	// Ambil tier element
	elementTier, exists := tierMap[element]
	if !exists {
		return models.RecipeTree{} // Return kosong jika elemen tidak ada
	}
	
	// Gunakan container/list untuk queue BFS yang proper
	queue := list.New()
	visited := make(map[string]bool)
	nodeMap := make(map[string]models.RecipeTree)
	
	// Inisialisasi dengan element yang dicari
	initialNode := models.RecipeTree{
		Root:     element,
		Left:     "",
		Right:    "",
		Tier:     fmt.Sprintf("%d", elementTier),
		Children: []models.RecipeTree{},
	}
	
	queue.PushBack(element)
	visited[element] = true
	nodeMap[element] = initialNode
	
	// Lakukan BFS
	for queue.Len() > 0 {
		current := queue.Remove(queue.Front()).(string)
		(*nodesVisited)++
		
		// Get current node
		currentNode := nodeMap[current]
		
		// Jika ini elemen dasar, selesai dengan branch ini
		if IsBasicElementWithTier(current, tierMap) {
			currentNode.Tier = "0"
			nodeMap[current] = currentNode
			continue
		}
		
		// Cari resep untuk elemen saat ini
		recipes, found := recipeMap[current]
		if !found || len(recipes) == 0 {
			continue
		}
		
		// Ambil tier elemen saat ini
		currentTier, tierExists := tierMap[current]
		if !tierExists {
			continue
		}
		
		// Cari resep valid
		foundValidRecipe := false
		
		for _, recipe := range recipes {
			if len(recipe) != 2 {
				continue
			}
			
			comp1 := recipe[0]
			comp2 := recipe[1]
			
			// Cek tier komponen
			comp1Tier, comp1Exists := tierMap[comp1]
			comp2Tier, comp2Exists := tierMap[comp2]
			
			if !comp1Exists || !comp2Exists {
				continue
			}
			
			// Cek constraint tier
			if comp1Tier >= currentTier || comp2Tier >= currentTier {
				continue
			}
			
			// Resep valid ditemukan, update node saat ini
			currentNode.Left = comp1
			currentNode.Right = comp2
			nodeMap[current] = currentNode
			
			// Tambahkan komponen ke queue jika belum dikunjungi
			if !visited[comp1] {
				queue.PushBack(comp1)
				visited[comp1] = true
				nodeMap[comp1] = models.RecipeTree{
					Root:     comp1,
					Left:     "",
					Right:    "",
					Tier:     fmt.Sprintf("%d", comp1Tier),
					Children: []models.RecipeTree{},
				}
			}
			
			if !visited[comp2] {
				queue.PushBack(comp2)
				visited[comp2] = true
				nodeMap[comp2] = models.RecipeTree{
					Root:     comp2,
					Left:     "",
					Right:    "",
					Tier:     fmt.Sprintf("%d", comp2Tier),
					Children: []models.RecipeTree{},
				}
			}
			
			foundValidRecipe = true
			break // Ambil resep valid pertama
		}
		
		// Jika tidak ada resep valid untuk node non-basic, ini path tidak valid
		if !foundValidRecipe && !IsBasicElementWithTier(current, tierMap) {
			return models.RecipeTree{} // Return kosong
		}
	}
	
	// Bangun tree lengkap dari nodeMap
	return buildCompleteTree(element, nodeMap)
}

// Membangun tree lengkap dari node map
func buildCompleteTree(root string, nodeMap map[string]models.RecipeTree) models.RecipeTree {
	result, exists := nodeMap[root]
	if !exists {
		return models.RecipeTree{} // Return kosong jika node tidak ada
	}
	
	// Jika basic element atau tidak punya children, langsung return
	if result.Left == "" || result.Right == "" {
		return result
	}
	
	// Rekursif build children
	leftChild := buildCompleteTree(result.Left, nodeMap)
	rightChild := buildCompleteTree(result.Right, nodeMap)
	
	// Jika salah satu child invalid, seluruh path invalid
	if leftChild.Root == "" || rightChild.Root == "" {
		return models.RecipeTree{}
	}
	
	// Tambahkan children ke result
	result.Children = []models.RecipeTree{leftChild, rightChild}
	
	return result
}

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