package search

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"container/list"
	"Tubes2_alchendol/models"
)

// MultipleBidirectional searches for multiple recipes using concurrent bidirectional search with multithreading
func MultipleBidirectional(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	// Cap maximum recipes at 50
	
	startTime := time.Now()
	
	// Filter elements based on tier constraint and Time exclusion
	elementMap, targetFound := CreateFilteredElementMap(elements, target)
	if !targetFound {
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
	
	// Get all top-level recipes for the target
	targetElements, found := elementMap[target]
	if !found || len(targetElements) == 0 {
		return []models.RecipeTree{}, time.Since(startTime).Seconds(), 0
	}
	
	// Shared state
	var results []models.RecipeTree
	var resultsMutex sync.Mutex
	recipeTreeMap := make(map[string]bool)
	
	// Track total nodes visited
	var totalNodesVisited int32
	
	// Channel to collect results
	resultChan := make(chan models.RecipeTree, maxRecipes*10)
	
	// Done channel for early stopping
	done := make(chan struct{})
	
	// WaitGroup for worker goroutines
	var wg sync.WaitGroup
	
	// Start result collector goroutine
	go func() {
		for tree := range resultChan {
			treeKey := generateTreeKey(tree)
			
			resultsMutex.Lock()
			// Check if we haven't seen this tree and haven't reached max
			if !recipeTreeMap[treeKey] && len(results) < maxRecipes {
				recipeTreeMap[treeKey] = true
				results = append(results, tree)
				
				// Signal early stop if we've reached exactly maxRecipes
				if len(results) == maxRecipes {
					close(done)
				}
			}
			resultsMutex.Unlock()
		}
	}()
	
	// Create recipe map for bidirectional search
	recipeMap := createRecipeMap(elements)
	
	// Process each top-level recipe in its own goroutine
	for i, element := range targetElements {
		if len(element.Recipes) != 2 {
			continue
		}
		
		comp1 := element.Recipes[0]
		comp2 := element.Recipes[1]
		
		wg.Add(1)
		go func(c1, c2 string, idx int) {
			defer wg.Done()
			
			// Check if we should stop
			select {
			case <-done:
				return
			default:
			}
			
			// Perform bidirectional search for this specific recipe
			bidirectionalWorkerConcurrent(target, c1, c2, elementMap, recipeMap,
				resultChan, done, &totalNodesVisited)
			
		}(comp1, comp2, i)
	}
	
	// Wait for all workers to complete
	wg.Wait()
	
	// Close result channel after all workers are done
	close(resultChan)
	
	// Give collector goroutine time to finish
	time.Sleep(10 * time.Millisecond)
	
	return results, time.Since(startTime).Seconds(), int(totalNodesVisited)
}

// bidirectionalWorkerConcurrent performs bidirectional search for all variations of a specific recipe
func bidirectionalWorkerConcurrent(target, comp1, comp2 string, elementMap map[string][]models.Element,
	recipeMap map[string][][]string, resultChan chan<- models.RecipeTree, done <-chan struct{}, 
	totalNodesVisited *int32) {
	
	// Create complete element map for easier access
	completeElementMap := make(map[string]models.Element)
	for name, elems := range elementMap {
		if len(elems) > 0 {
			completeElementMap[name] = elems[0]
		}
	}
	
	// Find all variations using bidirectional search
	variations := bidirectionalFindAllVariations(target, comp1, comp2, elementMap, 
		completeElementMap, recipeMap, done, totalNodesVisited)
	
	// Send all variations
	for _, tree := range variations {
		select {
		case <-done:
			return
		case resultChan <- tree:
		}
	}
}

// bidirectionalFindAllVariations finds all recipe variations using bidirectional search
func bidirectionalFindAllVariations(target, comp1, comp2 string, 
	elementMap map[string][]models.Element, completeElementMap map[string]models.Element,
	recipeMap map[string][][]string, done <-chan struct{}, totalNodesVisited *int32) []models.RecipeTree {
	
	var results []models.RecipeTree
	
	// Get variations for component 1
	comp1Trees := bidirectionalSearchVariations(comp1, elementMap, completeElementMap, 
		recipeMap, done, totalNodesVisited)
	
	// Get variations for component 2
	comp2Trees := bidirectionalSearchVariations(comp2, elementMap, completeElementMap,
		recipeMap, done, totalNodesVisited)
	
	// Combine all variations
	for _, comp1Tree := range comp1Trees {
		for _, comp2Tree := range comp2Trees {
			// Check if we should stop
			select {
			case <-done:
				return results
			default:
			}
			
			recipeTree := models.RecipeTree{
				Root:     target,
				Left:     comp1,
				Right:    comp2,
				Tier:     getTierAsString(target, elementMap),
				Children: []models.RecipeTree{comp1Tree, comp2Tree},
			}
			results = append(results, recipeTree)
		}
	}
	
	return results
}

// bidirectionalSearchVariations finds all variations for an element using bidirectional search
func bidirectionalSearchVariations(element string,elementMap map[string][]models.Element ,
	completeElementMap map[string]models.Element, recipeMap map[string][][]string,
	done <-chan struct{}, totalNodesVisited *int32) []models.RecipeTree {
	
	// Check if we should stop
	select {
	case <-done:
		return []models.RecipeTree{}
	default:
	}
	
	// If basic element, return directly
	if IsBasicElement(element) {
		atomic.AddInt32(totalNodesVisited, 1)
		return []models.RecipeTree{
			{
				Root:     element,
				Left:     "",
				Right:    "",
				Tier:     "0",
				Children: []models.RecipeTree{},
			},
		}
	}
	
	// Setup for bidirectional search
	forwardQueue := list.New()
	backwardQueue := list.New()
	
	forwardVisited := make(map[string]bool)
	backwardVisited := make(map[string]bool)
	
	forwardPaths := make(map[string][][]string) // Multiple paths per element
	backwardPaths := make(map[string][][]string)
	
	// Initialize forward search from basic elements
	basicElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, basic := range basicElements {
		forwardQueue.PushBack(basic)
		forwardVisited[basic] = true
		forwardPaths[basic] = [][]string{{basic}}
	}
	
	// Initialize backward search from target
	backwardQueue.PushBack(element)
	backwardVisited[element] = true
	backwardPaths[element] = [][]string{{element}}
	
	var meetingPoints []struct{
		Element string
		ForwardPaths [][]string
		BackwardPaths [][]string
	}
	
	// Perform bidirectional search
	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 {
		// Forward step
		if forwardQueue.Len() > 0 {
			forwardSize := forwardQueue.Len()
			for i := 0; i < forwardSize; i++ {
				current := forwardQueue.Remove(forwardQueue.Front()).(string)
				atomic.AddInt32(totalNodesVisited, 1)
				
				// Explore all recipes that can be made with current element
				for other := range forwardVisited {
					for result, recipes := range recipeMap {
						resultElem, exists := completeElementMap[result]
						if !exists {
							continue
						}
						
						currentElem := completeElementMap[current]
						otherElem := completeElementMap[other]
						
						// Check tier constraint
						if resultElem.Tier <= currentElem.Tier || resultElem.Tier <= otherElem.Tier {
							continue
						}
						
						for _, recipe := range recipes {
							if (recipe[0] == current && recipe[1] == other) ||
							   (recipe[1] == current && recipe[0] == other) {
								
								if !forwardVisited[result] {
									forwardVisited[result] = true
									forwardQueue.PushBack(result)
									
									// Create new paths
									var newPaths [][]string
									for _, path := range forwardPaths[current] {
										newPath := append([]string{}, path...)
										newPath = append(newPath, other, result)
										newPaths = append(newPaths, newPath)
									}
									forwardPaths[result] = newPaths
									
									// Check if we met with backward search
									if backwardVisited[result] {
										meetingPoints = append(meetingPoints, struct{
											Element string
											ForwardPaths [][]string
											BackwardPaths [][]string
										}{
											Element: result,
											ForwardPaths: forwardPaths[result],
											BackwardPaths: backwardPaths[result],
										})
									}
								}
							}
						}
					}
				}
			}
		}
		
		// Backward step
		if backwardQueue.Len() > 0 {
			backwardSize := backwardQueue.Len()
			for i := 0; i < backwardSize; i++ {
				current := backwardQueue.Remove(backwardQueue.Front()).(string)
				atomic.AddInt32(totalNodesVisited, 1)
				
				recipes, exists := recipeMap[current]
				if !exists {
					continue
				}
				
				currentElem := completeElementMap[current]
				
				for _, recipe := range recipes {
					for _, ingredient := range recipe {
						ingredientElem, exists := completeElementMap[ingredient]
						if !exists {
							continue
						}
						
						// Check tier constraint
						if currentElem.Tier <= ingredientElem.Tier {
							continue
						}
						
						if !backwardVisited[ingredient] {
							backwardVisited[ingredient] = true
							backwardQueue.PushBack(ingredient)
							
							// Create new paths
							var newPaths [][]string
							for _, path := range backwardPaths[current] {
								newPath := append([]string{}, path...)
								newPath = append(newPath, ingredient)
								newPaths = append(newPaths, newPath)
							}
							backwardPaths[ingredient] = newPaths
							
							// Check if we met with forward search
							if forwardVisited[ingredient] {
								meetingPoints = append(meetingPoints, struct{
									Element string
									ForwardPaths [][]string
									BackwardPaths [][]string
								}{
									Element: ingredient,
									ForwardPaths: forwardPaths[ingredient],
									BackwardPaths: backwardPaths[ingredient],
								})
							}
						}
					}
				}
			}
		}
	}
	
	// Build recipe trees from meeting points
	var results []models.RecipeTree
	uniqueResults := make(map[string]bool)
	
	for _, mp := range meetingPoints {
		for _, fPath := range mp.ForwardPaths {
			for _, bPath := range mp.BackwardPaths {
				// Build complete path
				tree := buildRecipeTreeFromPaths(element, mp.Element, fPath, bPath, 
					completeElementMap, recipeMap)
				
				key := generateTreeKey(tree)
				if !uniqueResults[key] {
					uniqueResults[key] = true
					results = append(results, tree)
				}
			}
		}
	}
	
	return results
}

// buildRecipeTreeFromPaths builds a recipe tree from forward and backward paths
func buildRecipeTreeFromPaths(target, meetingPoint string, forwardPath, backwardPath []string,
	completeElementMap map[string]models.Element, recipeMap map[string][][]string) models.RecipeTree {
	
	visited := make(map[string]bool)
	return buildRecipeTreeRecursive(target, visited, completeElementMap, recipeMap)
}

// buildRecipeTreeRecursive recursively builds the recipe tree
func buildRecipeTreeRecursive(element string, visited map[string]bool,
	completeElementMap map[string]models.Element, recipeMap map[string][][]string) models.RecipeTree {
	
	if visited[element] {
		return models.RecipeTree{
			Root:     element,
			Left:     "",
			Right:    "",
			Tier:     fmt.Sprintf("%d", completeElementMap[element].Tier),
			Children: []models.RecipeTree{},
		}
	}
	
	visited[element] = true
	
	tree := models.RecipeTree{
		Root:     element,
		Left:     "",
		Right:    "",
		Tier:     fmt.Sprintf("%d", completeElementMap[element].Tier),
		Children: []models.RecipeTree{},
	}
	
	// Basic element
	if IsBasicElement(element) {
		return tree
	}
	
	// Find valid recipe
	var bestRecipe []string
	lowestTierSum := 1000000
	
	recipes, exists := recipeMap[element]
	if exists {
		currentElem := completeElementMap[element]
		
		for _, recipe := range recipes {
			tierSum := 0
			valid := true
			
			for _, ingredient := range recipe {
				ingredientElem := completeElementMap[ingredient]
				tierSum += ingredientElem.Tier
				
				// Check tier constraint
				if currentElem.Tier <= ingredientElem.Tier {
					valid = false
					break
				}
			}
			
			if valid && tierSum < lowestTierSum {
				lowestTierSum = tierSum
				bestRecipe = recipe
			}
		}
	}
	
	if len(bestRecipe) == 0 {
		return tree
	}
	
	tree.Left = bestRecipe[0]
	tree.Right = bestRecipe[1]
	
	leftChild := buildRecipeTreeRecursive(bestRecipe[0], visited, completeElementMap, recipeMap)
	rightChild := buildRecipeTreeRecursive(bestRecipe[1], visited, completeElementMap, recipeMap)
	
	tree.Children = append(tree.Children, leftChild, rightChild)
	
	return tree
}

// Helper functions that already exist in your code

// generateTreeKey generates a unique key for a recipe tree
func generateTreeKey(tree models.RecipeTree) string {
	var keyBuilder strings.Builder
	generateTreeKeyHelper(tree, &keyBuilder)
	return keyBuilder.String()
}

func generateTreeKeyHelper(tree models.RecipeTree, sb *strings.Builder) {
	if len(tree.Children) == 0 {
		sb.WriteString(tree.Root)
		return
	}
	
	sb.WriteString(tree.Root)
	sb.WriteString("(")
	
	if len(tree.Children) >= 2 {
		leftKey := tree.Children[0].Root
		rightKey := tree.Children[1].Root
		
		if leftKey < rightKey {
			generateTreeKeyHelper(tree.Children[0], sb)
			sb.WriteString("+")
			generateTreeKeyHelper(tree.Children[1], sb)
		} else {
			generateTreeKeyHelper(tree.Children[1], sb)
			sb.WriteString("+")
			generateTreeKeyHelper(tree.Children[0], sb)
		}
	} else if len(tree.Children) == 1 {
		generateTreeKeyHelper(tree.Children[0], sb)
	}
	
	sb.WriteString(")")
}

// getTierAsString gets tier as string
func getTierAsString(element string, elementMap map[string][]models.Element) string {
	return fmt.Sprintf("%d", getTier(element, elementMap))
}

// getTier gets tier of an element
func getTier(element string, elementMap map[string][]models.Element) int {
	if IsBasicElement(element) {
		return 0
	}
	
	elements, found := elementMap[element]
	if !found || len(elements) == 0 {
		return -1
	}
	
	return elements[0].Tier
}