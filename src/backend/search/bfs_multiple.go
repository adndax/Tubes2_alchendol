package search

import (
	"Tubes2_alchendol/models"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

func processFilteredRecipes(recipes []models.Element, target string) (map[string][][]string, map[string]int) {
	foundTarget := false
	recipeMap := make(map[string][][]string)
	tierMap := make(map[string]int)

	// Simpan semua tier terlebih dahulu
	for _, r := range recipes {
		tierMap[r.Name] = r.Tier
	}

	// Untuk deteksi resep duplikat (kombinasi A+B = B+A)
	typeKeySet := make(map[string]map[string]bool)

	for _, r := range recipes {
		if len(r.Recipes) < 2 {
			continue
		}

		left := r.Recipes[0]
		right := r.Recipes[1]

		// Saring hanya resep dengan bahan yang tier-nya < tier hasil
		leftTier, leftOk := tierMap[left]
		rightTier, rightOk := tierMap[right]
		if !leftOk || !rightOk {
			continue
		}
		if leftTier >= r.Tier || rightTier >= r.Tier {
			continue // bahan terlalu tinggi, skip resep ini
		}

		// Cegah duplikat resep
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

		// Masukkan ke recipeMap
		recipeMap[r.Name] = append(recipeMap[r.Name], []string{left, right})

		if r.Name == target {
			// Jika elemen target ditemukan, tambahkan ke recipeMap
			foundTarget = true
		}

		if foundTarget {
			// Jika elemen target ditemukan, tambahkan ke recipeMap
			if r.Name != target {
				// Jika elemen target ditemukan, tambahkan ke recipeMap
				break
			}
		}
	}

	return recipeMap, tierMap
}

func MultipleBFS(target string, elements []models.Element, maxRecipes int) ([]TreeNode, float64, int) {
	start := time.Now()
	recipeMap, tierMap := processFilteredRecipes(elements, target)

	totalnodes := 0
	var results []TreeNode

	var mu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recipes, exists := recipeMap[target]
	if !exists {
		return nil, time.Since(start).Seconds(), 0
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

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

			root := &TreeNode{
				Root:     target,
				Left:     left,
				Right:    right,
				Tier:     fmt.Sprintf("%d", tierMap[target]),
				Children: []*TreeNode{},
			}

			nodeCount := 1 // mulai dari root
			type queueItem struct {
				parentNode *TreeNode
				element    string
				tier       int
			}

			queue := []queueItem{
				{root, left, leftTier},
				{root, right, rightTier},
			}

			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				nodeCount++

				if IsBasicElement(current.element) {
					leaf := &TreeNode{
						Root:     current.element,
						Left:     "",
						Right:    "",
						Tier:     "0",
						Children: nil,
					}
					current.parentNode.Children = append(current.parentNode.Children, leaf)
					continue
				}

				elementRecipes, exists := recipeMap[current.element]
				if !exists {
					continue
				}

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

					childNode := &TreeNode{
						Root:     current.element,
						Left:     left,
						Right:    right,
						Tier:     fmt.Sprintf("%d", current.tier),
						Children: []*TreeNode{},
					}
					current.parentNode.Children = append(current.parentNode.Children, childNode)
					queue = append(queue,
						queueItem{childNode, left, leftTier},
						queueItem{childNode, right, rightTier},
					)
					if current.element != target {
						break
					}
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if len(results) < maxRecipes {
				results = append(results, *root)
				totalnodes += nodeCount
				if len(results) >= maxRecipes {
					cancel()
				}
			}
		}(recipe)
	}

	wg.Wait()
	return results, time.Since(start).Seconds(), totalnodes
}
