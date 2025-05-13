package search

import (
	"Tubes2_alchendol/models"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"container/list"
)

func MultipleBFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	startTime := time.Now()
	
	// Tingkatkan timeout untuk elemen kompleks
	timeoutDuration := 120 * time.Second // Tingkatkan dari 15/30/60 ke 120 detik
	for _, el := range elements {
		if el.Name == target {
			// Complex elements get more time based on tier
			if el.Tier > 5 {
				timeoutDuration = 180 * time.Second // 3 menit untuk tier tinggi
			} else if el.Tier > 3 {
				timeoutDuration = 120 * time.Second // 2 menit untuk tier menengah
			}
			break
		}
	}
	
	// Create a context for coordinated cancellation
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()
	
	// Filter elements based on tier constraint 
	elementMap, targetFound := CreateFilteredElementMap(elements, target)
	if !targetFound {
		fmt.Printf("Target '%s' not found in filtered map\n", target)
		return []models.RecipeTree{}, time.Since(startTime).Seconds(), 0
	}
	
	// Add basic elements to map
	AddBasicElements(elementMap)
	
	// Check if target is a basic element
	if IsBasicElement(target) {
		basicTree := models.RecipeTree{
			Root:     target,
			Left:     "",
			Right:    "",
			Tier:     "0",
			Children: []models.RecipeTree{},
		}
		
		return []models.RecipeTree{basicTree}, time.Since(startTime).Seconds(), 1
	}
	
	// Gunakan atomic counter untuk node visited
	var totalNodesVisited int64 = 0
	
	// Hasil recipes dan map untuk menjaga keunikan
	var results []models.RecipeTree
	var uniqueRecipes = make(map[string]bool)
	var resultsMutex sync.Mutex
	
	// Get all direct recipes for the target
	targetElements, found := elementMap[target]
	if !found || len(targetElements) == 0 {
		fmt.Printf("No recipes found for target '%s' in elementMap\n", target)
		return []models.RecipeTree{}, time.Since(startTime).Seconds(), 0
	}
	
	// Kumpulkan semua kombinasi untuk target
	var allCombinations [][]string
	processedCombos := make(map[string]bool)
	
	for _, element := range targetElements {
		if len(element.Recipes) != 2 {
			continue
		}
		
		comp1 := element.Recipes[0]
		comp2 := element.Recipes[1]
		
		// Create a unique key for this combination
		var comboKey string
		if comp1 < comp2 {
			comboKey = comp1 + "+" + comp2
		} else {
			comboKey = comp2 + "+" + comp1
		}
		
		if processedCombos[comboKey] {
			continue
		}
		
		processedCombos[comboKey] = true
		allCombinations = append(allCombinations, []string{comp1, comp2})
	}
	
	fmt.Printf("Found %d unique direct combinations for %s\n", len(allCombinations), target)
	
	// Channel untuk mengumpulkan hasil - tingkatkan buffer size
	resultChan := make(chan models.RecipeTree, 1000)
	
	// Channel untuk sinyal done
	doneChan := make(chan struct{})
	
	// Mulai collector goroutine
	go func() {
		treesProcessed := 0
		for tree := range resultChan {
			if tree.Root == "" {
				continue
			}
			
			// Generate a unique key for this tree
			treeKey := generateBFSTreeKey(tree)
			
			resultsMutex.Lock()
			
			// Check if we've reached max recipes
			if len(results) >= maxRecipes {
				resultsMutex.Unlock()
				continue
			}
			
			// Only add if we haven't seen this tree before
			if !uniqueRecipes[treeKey] {
				uniqueRecipes[treeKey] = true
				results = append(results, tree)
				
				// Log progress dengan lebih efisien
				treesProcessed++
				if treesProcessed % 10 == 0 || treesProcessed < 10 {
					fmt.Printf("Added recipe %d/%d: %s = %s + %s (NodesVisited: %d)\n", 
						len(results), maxRecipes, tree.Root, tree.Left, tree.Right, 
						atomic.LoadInt64(&totalNodesVisited))
				}
				
				// If we've reached max recipes, signal done
				if len(results) >= maxRecipes {
					select {
					case doneChan <- struct{}{}:
					default:
					}
				}
			}
			
			resultsMutex.Unlock()
		}
	}()
	
	// Gunakan worker pool untuk pemrosesan paralel yang lebih efisien
	combinationChan := make(chan []string, len(allCombinations))
	var wg sync.WaitGroup
	
	// Jumlah worker berdasarkan CPU
	workerCount := runtime.NumCPU()
	fmt.Printf("Starting %d workers for processing combinations\n", workerCount)
	
	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			// Buat cache lokal untuk worker ini untuk menyimpan path yang ditemukan
			localPathCache := make(map[string]models.RecipeTree)
			
			for combo := range combinationChan {
				// Check for early termination
				select {
				case <-ctx.Done():
					return
				case <-doneChan:
					return
				default:
					// Continue processing
				}
				
				// Get component elements
				comp1 := combo[0]
				comp2 := combo[1]
				
				// Cek tier constraint
				targetTier := getElementTier(target, elementMap)
				comp1Tier := getElementTier(comp1, elementMap)
				comp2Tier := getElementTier(comp2, elementMap)
				
				if comp1Tier >= targetTier || comp2Tier >= targetTier {
					continue
				}
				
				// Proses kombinasi dengan cara BFS yang lebih efisien
				bfsProcessCombinationWithVariations(
					target, comp1, comp2, elementMap, 
					resultChan, ctx, doneChan, &totalNodesVisited, localPathCache)
			}
		}(i)
	}
	
	// Kirim kombinasi ke channel untuk diproses
	for _, combo := range allCombinations {
		combinationChan <- combo
	}
	close(combinationChan)
	
	// Tunggu semua worker selesai
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Tunggu sampai context done, atau doneChan signal, atau semua worker selesai
	select {
	case <-ctx.Done():
		fmt.Println("Context timeout reached, using results gathered so far")
	case <-doneChan:
		fmt.Printf("Found all %d requested recipes\n", maxRecipes)
	case <-func() chan struct{} {
		ch := make(chan struct{})
		go func() {
			wg.Wait()
			close(ch)
		}()
		return ch
	}():
		fmt.Println("All combinations processed")
	}
	
	// Wait a bit for the collector to finish
	cleanupTimeout := time.After(2 * time.Second)
	<-cleanupTimeout
	
	// Debug output if enabled
	if os.Getenv("DEBUG_RECIPES") == "1" && len(results) > 1 {
		DebugDisplayRecipeDifferences(results)
	}
	
	finalNodesVisited := atomic.LoadInt64(&totalNodesVisited)
	
	fmt.Printf("MultipleBFS returning %d recipes for %s with %d nodes visited\n", 
		len(results), target, finalNodesVisited)
	
	return results, time.Since(startTime).Seconds(), int(finalNodesVisited)
}

// Proses kombinasi dengan variasi menggunakan BFS
func bfsProcessCombinationWithVariations(
	target, comp1, comp2 string, 
	elementMap map[string][]models.Element,
	resultChan chan<- models.RecipeTree,
	ctx context.Context,
	doneChan <-chan struct{},
	totalNodesVisited *int64,
	pathCache map[string]models.RecipeTree) {
	
	// Increment visit counter
	atomic.AddInt64(totalNodesVisited, 1)
	
	// Get target tier
	targetTier := getElementTier(target, elementMap)
	
	// Cari variasi untuk komponen pertama dengan BFS yang efisien
	comp1Variations := findBFSComponentVariations(
		comp1, elementMap, ctx, doneChan, totalNodesVisited, pathCache)
	
	// Check for termination
	select {
	case <-ctx.Done():
		return
	case <-doneChan:
		return
	default:
	}
	
	// Cari variasi untuk komponen kedua dengan BFS yang efisien
	comp2Variations := findBFSComponentVariations(
		comp2, elementMap, ctx, doneChan, totalNodesVisited, pathCache)
	
	// Check for termination
	select {
	case <-ctx.Done():
		return
	case <-doneChan:
		return
	default:
	}
	
	// Batasi jumlah kombinasi untuk mencegah ledakan
	maxCombosPerPair := 200
	combinationCount := 0
	
	// Kombinasikan semua variasi komponen untuk membuat resep lengkap
	for _, v1 := range comp1Variations {
		// Cek terminasi secara periodik
		if combinationCount % 10 == 0 {
			select {
			case <-ctx.Done():
				return
			case <-doneChan:
				return
			default:
			}
		}
		
		for _, v2 := range comp2Variations {
			// Buat resep tree dengan variasi ini
			recipeTree := models.RecipeTree{
				Root:     target,
				Left:     comp1,
				Right:    comp2,
				Tier:     fmt.Sprintf("%d", targetTier),
				Children: []models.RecipeTree{v1, v2},
			}
			
			// Kirim ke channel
			select {
			case resultChan <- recipeTree:
				atomic.AddInt64(totalNodesVisited, 1)
			case <-ctx.Done():
				return
			case <-doneChan:
				return
			default:
				// Channel full, skip
			}
			
			combinationCount++
			if combinationCount >= maxCombosPerPair {
				break
			}
		}
		
		if combinationCount >= maxCombosPerPair {
			break
		}
	}
}

// Temukan variasi komponen dengan BFS yang efisien
func findBFSComponentVariations(
	element string,
	elementMap map[string][]models.Element,
	ctx context.Context,
	doneChan <-chan struct{},
	totalNodesVisited *int64,
	pathCache map[string]models.RecipeTree) []models.RecipeTree {
	
	// Increment visit counter
	atomic.AddInt64(totalNodesVisited, 1)
	
	// Check for termination
	select {
	case <-ctx.Done():
		return []models.RecipeTree{}
	case <-doneChan:
		return []models.RecipeTree{}
	default:
	}
	
	// Cek cache untuk element ini
	if cachedTree, found := pathCache[element]; found {
		return []models.RecipeTree{cachedTree}
	}
	
	// Jika element adalah basic, langsung return
	if IsBasicElement(element) {
		basicTree := models.RecipeTree{
			Root:     element,
			Left:     "",
			Right:    "",
			Tier:     "0",
			Children: []models.RecipeTree{},
		}
		
		pathCache[element] = basicTree
		return []models.RecipeTree{basicTree}
	}
	
	// Get element tier
	elementTier := getElementTier(element, elementMap)
	
	// Gunakan BFS untuk temukan variasi resep
	var variations []models.RecipeTree
	
	// Gunakan batas yang tinggi - tidak terbatas
	maxVariationsPerElement := 100 // Tingkatkan dari 5 ke 100
	
	// Get recipes untuk element ini
	elementRecipes, found := elementMap[element]
	if !found || len(elementRecipes) == 0 {
		return []models.RecipeTree{}
	}
	
	// Track recipe combinations yang sudah diproses
	processedCombos := make(map[string]bool)
	
	// Proses setiap resep
	for _, recipe := range elementRecipes {
		// Cek jika sudah mencapai batas variasi
		if len(variations) >= maxVariationsPerElement {
			break
		}
		
		if len(recipe.Recipes) != 2 {
			continue
		}
		
		subComp1 := recipe.Recipes[0]
		subComp2 := recipe.Recipes[1]
		
		// Buat key unik untuk kombinasi ini
		var comboKey string
		if subComp1 < subComp2 {
			comboKey = subComp1 + "+" + subComp2
		} else {
			comboKey = subComp2 + "+" + subComp1
		}
		
		// Skip jika sudah diproses
		if processedCombos[comboKey] {
			continue
		}
		processedCombos[comboKey] = true
		
		// Cek tier constraint
		subComp1Tier := getElementTier(subComp1, elementMap)
		subComp2Tier := getElementTier(subComp2, elementMap)
		
		if subComp1Tier >= elementTier || subComp2Tier >= elementTier {
			continue
		}
		
		// Dapatkan variasi untuk subcomponents dengan BFS
		subComp1Variations := findBFSComponentVariations(
			subComp1, elementMap, ctx, doneChan, totalNodesVisited, pathCache)
		
		// Check for termination
		select {
		case <-ctx.Done():
			return variations
		case <-doneChan:
			return variations
		default:
		}
		
		// Skip jika tidak ada variasi
		if len(subComp1Variations) == 0 {
			continue
		}
		
		subComp2Variations := findBFSComponentVariations(
			subComp2, elementMap, ctx, doneChan, totalNodesVisited, pathCache)
		
		// Check for termination
		select {
		case <-ctx.Done():
			return variations
		case <-doneChan:
			return variations
		default:
		}
		
		// Skip jika tidak ada variasi
		if len(subComp2Variations) == 0 {
			continue
		}
		
		// Batasi kombinasi per resep untuk menghindari ledakan
		maxCombosPerRecipe := 20
		combosAdded := 0
		
		// Kombinasikan variasi subkomponen
		for _, v1 := range subComp1Variations {
			if len(variations) >= maxVariationsPerElement {
				break
			}
			
			for _, v2 := range subComp2Variations {
				if len(variations) >= maxVariationsPerElement {
					break
				}
				
				if combosAdded >= maxCombosPerRecipe {
					break
				}
				
				// Buat tree dengan variasi ini
				tree := models.RecipeTree{
					Root:     element,
					Left:     subComp1,
					Right:    subComp2,
					Tier:     fmt.Sprintf("%d", elementTier),
					Children: []models.RecipeTree{v1, v2},
				}
				
				variations = append(variations, tree)
				combosAdded++
			}
			
			if combosAdded >= maxCombosPerRecipe {
				break
			}
		}
	}
	
	// Cache hasil jika tidak kosong
	if len(variations) > 0 {
		pathCache[element] = variations[0]
	}
	
	return variations
}

// Implementasi bfsSearchComponent yang efisien menggunakan BFS untuk menentukan satu path
// dan menyimpan hasilnya di cache
func efficientBfsSearchComponent(
	element string, 
	elementMap map[string][]models.Element, 
	totalNodesVisited *int64,
	pathCache map[string]models.RecipeTree) models.RecipeTree {
	
	// Increment visit counter
	atomic.AddInt64(totalNodesVisited, 1)
	
	// Check cache
	if cachedTree, found := pathCache[element]; found {
		return cachedTree
	}
	
	// If element is basic, return directly
	if IsBasicElement(element) {
		basicTree := models.RecipeTree{
			Root:     element,
			Left:     "",
			Right:    "",
			Tier:     "0",
			Children: []models.RecipeTree{},
		}
		
		pathCache[element] = basicTree
		return basicTree
	}
	
	// Get element tier
	elementTier := getElementTier(element, elementMap)
	
	// Use BFS to find a path to basic elements
	queue := list.New()
	visited := make(map[string]bool)
	nodeMap := make(map[string]models.RecipeTree)
	
	// Create initial node for the element
	initialNode := models.RecipeTree{
		Root:     element,
		Left:     "",
		Right:    "",
		Tier:     fmt.Sprintf("%d", elementTier),
		Children: []models.RecipeTree{},
	}
	
	// Initialize with the target element
	queue.PushBack(element)
	visited[element] = true
	nodeMap[element] = initialNode
	
	// Process the queue
	for queue.Len() > 0 {
		// Get next element from queue
		current := queue.Remove(queue.Front()).(string)
		atomic.AddInt64(totalNodesVisited, 1)
		
		// Get current node
		currentNode := nodeMap[current]
		
		// If this is a basic element, we're done with this branch
		if IsBasicElement(current) {
			currentNode.Tier = "0"
			nodeMap[current] = currentNode
			continue
		}
		
		// Get recipes for current element
		elements, found := elementMap[current]
		if !found || len(elements) == 0 {
			continue
		}
		
		// Find a valid recipe
		currentTier := getElementTier(current, elementMap)
		foundRecipe := false
		
		for _, el := range elements {
			if len(el.Recipes) != 2 {
				continue
			}
			
			comp1 := el.Recipes[0]
			comp2 := el.Recipes[1]
			
			comp1Tier := getElementTier(comp1, elementMap)
			comp2Tier := getElementTier(comp2, elementMap)
			
			// Check tier constraints
			if comp1Tier >= currentTier || comp2Tier >= currentTier {
				continue
			}
			
			// Found a valid recipe
			currentNode.Left = comp1
			currentNode.Right = comp2
			nodeMap[current] = currentNode
			
			// Add components to queue if not visited
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
			
			foundRecipe = true
			break // Take first valid recipe
		}
		
		if !foundRecipe {
			// If no valid recipe found for a non-basic element, this path is invalid
			if !IsBasicElement(current) {
				return models.RecipeTree{}
			}
		}
	}
	
	// Build complete tree from nodeMap
	result := buildTreeFromNodeMap(element, nodeMap)
	
	// Cache the result
	if result.Root != "" {
		pathCache[element] = result
	}
	
	return result
}

// Build a complete tree from the node map
func buildTreeFromNodeMap(root string, nodeMap map[string]models.RecipeTree) models.RecipeTree {
	result, exists := nodeMap[root]
	if !exists {
		return models.RecipeTree{}
	}
	
	// If this is a basic element or has no children, return as is
	if result.Left == "" || result.Right == "" {
		return result
	}
	
	// Recursively build children
	leftChild := buildTreeFromNodeMap(result.Left, nodeMap)
	rightChild := buildTreeFromNodeMap(result.Right, nodeMap)
	
	// Skip if any child is invalid
	if leftChild.Root == "" || rightChild.Root == "" {
		return models.RecipeTree{}
	}
	
	// Add children to result
	result.Children = []models.RecipeTree{leftChild, rightChild}
	
	return result
}

// Helper function to generate a BFS-style key for a tree
func generateBFSTreeKey(tree models.RecipeTree) string {
	var sb strings.Builder
	generateBFSTreeKeyHelper(tree, &sb, 0)
	return sb.String()
}

func generateBFSTreeKeyHelper(tree models.RecipeTree, sb *strings.Builder, depth int) {
	// Limit recursion depth
	if depth > 20 {
		sb.WriteString(tree.Root)
		return
	}
	
	sb.WriteString(tree.Root)
	
	if len(tree.Children) == 0 {
		return
	}
	
	sb.WriteString("(")
	
	if len(tree.Children) >= 2 {
		// Process children in deterministic order
		left := tree.Children[0]
		right := tree.Children[1]
		
		generateBFSTreeKeyHelper(left, sb, depth+1)
		sb.WriteString("+")
		generateBFSTreeKeyHelper(right, sb, depth+1)
	} else if len(tree.Children) == 1 {
		generateBFSTreeKeyHelper(tree.Children[0], sb, depth+1)
	}
	
	sb.WriteString(")")
}