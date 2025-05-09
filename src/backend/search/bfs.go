package search

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Recipe struct {
	Name    string   `json:"name"`
	Recipes []string `json:"recipes"`
	Tier    int      `json:"tier"`
}

type TreeNode struct {
	Root     string     `json:"root"`
	Left     string     `json:"left"`
	Right    string     `json:"right"`
	Tier     string     `json:"tier"`
	Children []*TreeNode `json:"children"`
}

func loadRecipes(filename string) ([]Recipe, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var recipes []Recipe
	err = json.Unmarshal(data, &recipes)
	return recipes, err
}

func BFS(target string, targetTier int, recipeMap map[string][][]string, tierMap map[string]int) []TreeNode {
	var trees []TreeNode

	recipes, exists := recipeMap[target]
	if !exists {
		return trees
	}

	type result struct {
		tree TreeNode
		ok   bool
	}

	sem := make(chan struct{}, 5) // semaphore

	resultCh := make(chan result, len(recipes))

	for _, recipe := range recipes {
		// Jalankan satu goroutine per resep
		sem <- struct{}{}
		go func(recipe []string) {
			defer func() { <-sem }() // Lepaskan semaphore setelah selesai

			if len(recipe) != 2 {
				resultCh <- result{ok: false}
				return
			}

			left, right := recipe[0], recipe[1]
			leftTier, leftExists := tierMap[left]
			rightTier, rightExists := tierMap[right]

			if !leftExists || !rightExists || leftTier >= targetTier || rightTier > targetTier {
				resultCh <- result{ok: false}
				return
			}

			root := &TreeNode{
				Root:     target,
				Left:     left,
				Right:    right,
				Tier:     fmt.Sprintf("%d", targetTier),
				Children: []*TreeNode{},
			}

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

				if current.tier == 0 {
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
				}
			}

			resultCh <- result{tree: *root, ok: true}
		}(recipe)
	}

	for i := 0; i < len(recipes); i++ {
		res := <-resultCh
		if res.ok {
			trees = append(trees, res.tree)
		}
	}

	return trees
}


func processRecipes(recipes []Recipe, target string) (map[string][][]string, map[string]int) {
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

		// Ambil dua bahan pertama
		left := r.Recipes[0]
		right := r.Recipes[1]

		// Urutkan untuk deteksi duplikat: A+B sama dengan B+A
		pair := []string{left, right}
		sort.Strings(pair)
		recipeKey := pair[0] + "+" + pair[1]

		// Buat map untuk menyimpan kombinasi unik per elemen hasil
		if _, ok := typeKeySet[r.Name]; !ok {
			typeKeySet[r.Name] = make(map[string]bool)
		}

		// Skip jika sudah ada
		if typeKeySet[r.Name][recipeKey] {
			continue
		}

		// Tandai kombinasi ini sudah digunakan
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


func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Masukkan nama file JSON (default: output.json): ")
	fileInput, _ := reader.ReadString('\n')
	fileInput = strings.TrimSpace(fileInput)
	if fileInput == "" {
		fileInput = "output.json"
	}

	fmt.Print("Masukkan elemen target: ")
	targetInput, _ := reader.ReadString('\n')
	targetInput = strings.TrimSpace(targetInput)

	recipes, err := loadRecipes(fileInput)
	if err != nil {
		fmt.Println("Gagal membaca file:", err)
		return
	}

	if len(recipes) == 0 {
		fmt.Println("Tidak ada resep yang ditemukan dalam file")
		return
	}

	recipeMap, tierMap := processRecipes(recipes, targetInput)


	targetTier, exists := tierMap[targetInput]
	if !exists {
		fmt.Printf("Tier untuk elemen '%s' tidak ditemukan\n", targetInput)
		return
	}

	trees := BFS(targetInput, targetTier, recipeMap, tierMap)


	if len(trees) == 0 {
		fmt.Printf("Tidak ada resep yang ditemukan untuk elemen '%s' yang memenuhi batasan tier\n", targetInput)
		return
	}

	output, err := json.MarshalIndent(trees, "", "  ")
	if err != nil {
		fmt.Println("Gagal mengubah ke JSON:", err)
		return
	}

	fmt.Println(string(output))

	outputFile := fmt.Sprintf("%s_tree.json", targetInput)
	err = os.WriteFile(outputFile, output, 0644)
	if err != nil {
		fmt.Println("Gagal menulis ke file:", err)
		return
	}
	
	fmt.Printf("Pohon resep berhasil disimpan ke %s\n", outputFile)
}