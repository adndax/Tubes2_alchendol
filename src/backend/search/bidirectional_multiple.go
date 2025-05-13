package search

import (
	"Tubes2_alchendol/models"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MultipleBidirectional searches for multiple recipes using concurrent bidirectional search with multithreading
func MultipleBidirectional(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	// Cap maximum recipes at a reasonable number
	if maxRecipes > 50 {
		maxRecipes = 50
	}

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
	uniqueCompleteRecipes := make(map[string]bool) // For tracking complete unique trees

	// Track total nodes visited
	var totalNodesVisited int32

	// Channel to collect results
	resultChan := make(chan models.RecipeTree, maxRecipes*10)

	// Done channel for early stopping
	done := make(chan struct{})

	// WaitGroup for worker goroutines
	var wg sync.WaitGroup

	// Start result collector goroutine - improved to check complete tree structure
	go func() {
		for tree := range resultChan {
			// Generate a unique key for the entire tree, not just top level
			completeTreeKey := generateCompleteTreeKey(tree)

			resultsMutex.Lock()
			// Check if we haven't seen this exact tree structure and haven't reached max
			if !uniqueCompleteRecipes[completeTreeKey] && len(results) < maxRecipes {
				uniqueCompleteRecipes[completeTreeKey] = true
				results = append(results, tree)

				// Debug output showing we added a recipe
				fmt.Printf("Added recipe %d/%d for %s (unique key: %s)\n", 
					len(results), maxRecipes, tree.Root, completeTreeKey[:20]+"...")

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

	// Create a complete element map for easier access
	completeElementMap := make(map[string]models.Element)
	for name, elems := range elementMap {
		if len(elems) > 0 {
			completeElementMap[name] = elems[0]
		}
	}

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

			// Process this specific recipe combination with variations at all levels
			processRecipeCombinationWithMultipleVariations(target, c1, c2, elementMap, completeElementMap,
				recipeMap, resultChan, done, &totalNodesVisited)

		}(comp1, comp2, i)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Close result channel after all workers are done
	close(resultChan)

	// Give collector goroutine time to finish
	time.Sleep(10 * time.Millisecond)

	// Sort results by recipe complexity (optional)
	if len(results) > 1 {
		sort.Slice(results, func(i, j int) bool {
			// Sort by total nodes in the tree or any other criteria
			return countTreeNodes(results[i]) < countTreeNodes(results[j])
		})
	}

	fmt.Printf("MultipleBidirectional returning %d unique recipes with %d nodes visited\n", 
		len(results), int(totalNodesVisited))

	return results, time.Since(startTime).Seconds(), int(totalNodesVisited)
}

// Count nodes in a tree for sorting
func countTreeNodes(tree models.RecipeTree) int {
	if len(tree.Children) == 0 {
		return 1
	}
	
	count := 1 // Count this node
	for _, child := range tree.Children {
		count += countTreeNodes(child)
	}
	return count
}

// processRecipeCombinationWithMultipleVariations processes a recipe combination with variations at all levels
func processRecipeCombinationWithMultipleVariations(target, comp1, comp2 string, 
	elementMap map[string][]models.Element, completeElementMap map[string]models.Element,
	recipeMap map[string][][]string, resultChan chan<- models.RecipeTree, 
	done <-chan struct{}, totalNodesVisited *int32) {

	// Check tier constraints before proceeding
	targetTier := completeElementMap[target].Tier
	comp1Tier := -1
	comp2Tier := -1
	
	if el, exists := completeElementMap[comp1]; exists {
		comp1Tier = el.Tier
	} else if IsBasicElement(comp1) {
		comp1Tier = 0
	}
	
	if el, exists := completeElementMap[comp2]; exists {
		comp2Tier = el.Tier
	} else if IsBasicElement(comp2) {
		comp2Tier = 0
	}
	
	// Skip tier violations
	if comp1Tier >= targetTier || comp2Tier >= targetTier {
		return
	}
	
	// Process different recipe variations for component 1
	comp1Trees := findAllRecipeVariations(comp1, elementMap, completeElementMap, 
		recipeMap, done, totalNodesVisited, make(map[string]bool))
	
	// Process different recipe variations for component 2
	comp2Trees := findAllRecipeVariations(comp2, elementMap, completeElementMap, 
		recipeMap, done, totalNodesVisited, make(map[string]bool))
	
	// Combine component variations to create full recipes
	for _, comp1Tree := range comp1Trees {
		// Check if we should stop
		select {
		case <-done:
			return
		default:
		}
		
		for _, comp2Tree := range comp2Trees {
			// Check if we should stop
			select {
			case <-done:
				return
			default:
			}
			
			// Create a complete recipe tree with these component variations
			recipeTree := models.RecipeTree{
				Root:     target,
				Left:     comp1,
				Right:    comp2,
				Tier:     fmt.Sprintf("%d", targetTier),
				Children: []models.RecipeTree{comp1Tree, comp2Tree},
			}
			
			// Send this variation to the result channel
			select {
			case resultChan <- recipeTree:
				// Increment counter whenever we send a variation
				atomic.AddInt32(totalNodesVisited, 1)
			case <-done:
				return
			}
		}
	}
}

// findAllRecipeVariations finds all valid recipe variations for an element recursively
func findAllRecipeVariations(element string, elementMap map[string][]models.Element,
	completeElementMap map[string]models.Element, recipeMap map[string][][]string,
	done <-chan struct{}, totalNodesVisited *int32, visited map[string]bool) []models.RecipeTree {
	
	// Check for early termination
	select {
	case <-done:
		return []models.RecipeTree{}
	default:
	}
	
	// Count this visit
	atomic.AddInt32(totalNodesVisited, 1)
	
	// Prevent cycles - if we've visited this element in this path, return empty
	if visited[element] {
		return []models.RecipeTree{}
	}
	
	// If this is a basic element, return it directly
	if IsBasicElement(element) {
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
	
	// Create a copy of the visited map for this path
	pathVisited := make(map[string]bool)
	for k, v := range visited {
		pathVisited[k] = v
	}
	pathVisited[element] = true
	
	// Get element tier
	elementTier := completeElementMap[element].Tier
	
	// Get all direct recipe combinations for this element
	elementRecipes, found := elementMap[element]
	if !found || len(elementRecipes) == 0 {
		return []models.RecipeTree{}
	}
	
	var results []models.RecipeTree
	processedCombos := make(map[string]bool)
	
	// For each recipe of this element
	for _, recipe := range elementRecipes {
		if len(recipe.Recipes) != 2 {
			continue
		}
		
		subComp1 := recipe.Recipes[0]
		subComp2 := recipe.Recipes[1]
		
		// Create a unique key for this combination
		var comboKey string
		if subComp1 < subComp2 {
			comboKey = subComp1 + "+" + subComp2
		} else {
			comboKey = subComp2 + "+" + subComp1
		}
		
		// Skip if we've already processed this combination
		if processedCombos[comboKey] {
			continue
		}
		processedCombos[comboKey] = true
		
		// Get component tiers
		subComp1Tier := -1
		if el, exists := completeElementMap[subComp1]; exists {
			subComp1Tier = el.Tier
		} else if IsBasicElement(subComp1) {
			subComp1Tier = 0
		}
		
		subComp2Tier := -1
		if el, exists := completeElementMap[subComp2]; exists {
			subComp2Tier = el.Tier
		} else if IsBasicElement(subComp2) {
			subComp2Tier = 0
		}
		
		// Check tier constraints
		if subComp1Tier >= elementTier || subComp2Tier >= elementTier {
			continue
		}
		
		// Recursively find variations for each component
		subComp1Trees := findAllRecipeVariations(subComp1, elementMap, completeElementMap, 
			recipeMap, done, totalNodesVisited, pathVisited)
		
		subComp2Trees := findAllRecipeVariations(subComp2, elementMap, completeElementMap, 
			recipeMap, done, totalNodesVisited, pathVisited)
		
		// If either component has no valid recipes, skip this combination
		if len(subComp1Trees) == 0 || len(subComp2Trees) == 0 {
			continue
		}
		
		// Combine all component variations to create recipe variations
		for _, comp1Tree := range subComp1Trees {
			for _, comp2Tree := range subComp2Trees {
				// Create a recipe tree with these component variations
				recipeTree := models.RecipeTree{
					Root:     element,
					Left:     subComp1,
					Right:    subComp2,
					Tier:     fmt.Sprintf("%d", elementTier),
					Children: []models.RecipeTree{comp1Tree, comp2Tree},
				}
				
				// Add to results
				results = append(results, recipeTree)
				
				// Limit the number of variations per element to avoid explosion
				// Only keep at most 5 variations per element to manage complexity
				if len(results) >= 5 {
					return results
				}
			}
		}
	}
	
	return results
}

// getTierAsString gets tier as string from element map
func getTierAsString(element string, elementMap map[string][]models.Element) string {
	elements, exists := elementMap[element]
	if !exists || len(elements) == 0 {
		return "0"
	}
	return fmt.Sprintf("%d", elements[0].Tier)
}

// generateTreeKey creates a simple key based on just the top level of tree
func generateTreeKey(tree models.RecipeTree) string {
	return tree.Root + "|" + tree.Left + "|" + tree.Right
}