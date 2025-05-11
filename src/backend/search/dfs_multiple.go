package search

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"Tubes2_alchendol/models"
)

// MultipleDFS searches for multiple recipes using concurrent DFS with multithreading
func MultipleDFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
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
			
			// Generate all combinations for this top-level recipe
			generateRecipeCombinations(target, c1, c2, elementMap, 
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

// generateRecipeCombinations generates all combinations for a specific recipe
func generateRecipeCombinations(target, comp1, comp2 string, elementMap map[string][]models.Element,
	resultChan chan<- models.RecipeTree, done <-chan struct{}, totalNodesVisited *int32) {
	
	// Get all complete variations for both components
	var comp1Trees, comp2Trees []models.RecipeTree
	var wg sync.WaitGroup
	
	wg.Add(2)
	
	// Get variations for comp1
	go func() {
		defer wg.Done()
		comp1Trees = getAllVariationsWithEarlyStopping(comp1, elementMap, []string{target}, 
			done, totalNodesVisited)
	}()
	
	// Get variations for comp2
	go func() {
		defer wg.Done()
		comp2Trees = getAllVariationsWithEarlyStopping(comp2, elementMap, []string{target}, 
			done, totalNodesVisited)
	}()
	
	wg.Wait()
	
	// Increment for this node
	atomic.AddInt32(totalNodesVisited, 1)
	
	// Generate all combinations
	for _, comp1Tree := range comp1Trees {
		for _, comp2Tree := range comp2Trees {
			// Check if we should stop
			select {
			case <-done:
				return
			default:
			}
			
			recipeTree := models.RecipeTree{
				Root:     target,
				Left:     comp1,
				Right:    comp2,
				Tier:     getTierAsString(target, elementMap),
				Children: []models.RecipeTree{comp1Tree, comp2Tree},
			}
			
			// Send result - non-blocking to avoid deadlock
			select {
			case resultChan <- recipeTree:
			case <-done:
				return
			}
		}
	}
}

// getAllVariationsWithEarlyStopping finds all variations with early stopping capability
func getAllVariationsWithEarlyStopping(element string, elementMap map[string][]models.Element, 
	visited []string, done <-chan struct{}, totalNodesVisited *int32) []models.RecipeTree {
	
	// Check if we should stop
	select {
	case <-done:
		return []models.RecipeTree{}
	default:
	}
	
	// Increment nodes visited
	atomic.AddInt32(totalNodesVisited, 1)
	
	// Check for cycles
	for _, v := range visited {
		if v == element {
			return []models.RecipeTree{}
		}
	}
	
	// Base case: basic element (tier 0)
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
	
	// Get all recipes for this element
	elements, found := elementMap[element]
	if !found || len(elements) == 0 {
		return []models.RecipeTree{}
	}
	
	var allCompleteTrees []models.RecipeTree
	
	// Explore each recipe
	for _, el := range elements {
		// Check if we should stop
		select {
		case <-done:
			return allCompleteTrees
		default:
		}
		
		if len(el.Recipes) != 2 {
			continue
		}
		
		comp1 := el.Recipes[0]
		comp2 := el.Recipes[1]
		
		// Check tier constraints
		comp1Tier := getTier(comp1, elementMap)
		comp2Tier := getTier(comp2, elementMap)
		elementTier := getTier(element, elementMap)
		
		if comp1Tier >= elementTier || comp2Tier >= elementTier {
			continue
		}
		
		// Get all complete variations for both components recursively
		comp1Trees := getAllVariationsWithEarlyStopping(comp1, elementMap, append(visited, element),
			done, totalNodesVisited)
		comp2Trees := getAllVariationsWithEarlyStopping(comp2, elementMap, append(visited, element),
			done, totalNodesVisited)
		
		// Only create combinations if both components have valid complete paths
		if len(comp1Trees) > 0 && len(comp2Trees) > 0 {
			for _, comp1Tree := range comp1Trees {
				for _, comp2Tree := range comp2Trees {
					// Check if we should stop
					select {
					case <-done:
						return allCompleteTrees
					default:
					}
					
					recipeTree := models.RecipeTree{
						Root:     element,
						Left:     comp1,
						Right:    comp2,
						Tier:     getTierAsString(element, elementMap),
						Children: []models.RecipeTree{comp1Tree, comp2Tree},
					}
					allCompleteTrees = append(allCompleteTrees, recipeTree)
				}
			}
		}
	}
	
	return allCompleteTrees
}

// Helper functions
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

func getTierAsString(element string, elementMap map[string][]models.Element) string {
	return fmt.Sprintf("%d", getTier(element, elementMap))
}