package search

import (
	"Tubes2_alchendol/models"
	"context"
	"sync"
	"time"
)

// MultipleBFS mengembalikan []models.RecipeTree untuk multiple recipes
func MultipleBFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	start := time.Now()
	recipeMap, tierMap := processRecipes(elements, target)

	totalnodes := 0
	var results []models.RecipeTree

	// Jika target adalah elemen dasar, langsung buat tree tunggal
	if IsBasicElement(target) {
		basicTree := createBasicElementLeaf(target)
		return []models.RecipeTree{basicTree}, time.Since(start).Seconds(), 1
	}

	var mu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recipes, exists := recipeMap[target]
	if !exists {
		return nil, time.Since(start).Seconds(), 0
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Semaphore untuk membatasi goroutine concurrent

	for _, recipe := range recipes {
		mu.Lock()
		if len(results) >= maxRecipes {
			mu.Unlock()
			break
		}
		mu.Unlock()

		sem <- struct{}{}
		wg.Add(1)
		go func(recipe []string) {
			defer func() {
				<-sem
				wg.Done()
			}()

			select {
			case <-ctx.Done():
				return
			default:
			}

			if len(recipe) != 2 {
				return
			}

			left, right := recipe[0], recipe[1]
			leftTier, leftExists := tierMap[left]
			rightTier, rightExists := tierMap[right]

			if !leftExists || !rightExists || leftTier >= tierMap[target] || rightTier > tierMap[target] {
				return
			}

			// Membuat root untuk resep ini
			root := createRecipeTreeNode(target, left, right, tierMap[target])

			// Slice untuk menyimpan semua node
			nodes := []models.RecipeTree{root}
			rootIdx := 0

			nodeCount := 1 // mulai dari root
			type queueItem struct {
				parentNodeIdx int    // Index dari node parent dalam slice nodes
				element       string
				tier          int
			}

			// Inisialisasi queue dengan 2 children dari root
			queue := []queueItem{
				{rootIdx, left, leftTier},
				{rootIdx, right, rightTier},
			}

			// BFS untuk membuat tree
			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				nodeCount++

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
					
					if !leftOk || !rightOk || leftTier > current.tier || rightTier > current.tier {
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
					
					// Untuk non-target element, cukup ambil 1 resep saja
					if current.element != target {
						break
					}
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if len(results) < maxRecipes {
				// Deep copy untuk menghindari referensi yang sama
				result := deepCopyRecipeTree(nodes[0])
				results = append(results, result)
				totalnodes += nodeCount
				
				// Jika sudah mencapai batas, batalkan goroutine lain
				if len(results) >= maxRecipes {
					cancel()
				}
			}
		}(recipe)
	}

	wg.Wait()
	return results, time.Since(start).Seconds(), totalnodes
}