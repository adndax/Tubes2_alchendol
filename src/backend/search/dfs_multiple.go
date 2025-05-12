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
	// Cap maximum recipes at 20 to prevent excessive combinations
	if maxRecipes > 20 {
		maxRecipes = 20
	}
	
	startTime := time.Now()
	
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
	
	// Get all top-level recipes for the target
	targetElements, found := elementMap[target]
	if !found || len(targetElements) == 0 {
		fmt.Printf("No recipes found for target '%s' in elementMap\n", target)
		return []models.RecipeTree{}, time.Since(startTime).Seconds(), 0
	}
	
	fmt.Printf("Found %d recipes for target '%s'\n", len(targetElements), target)
	
	// Use single DFS first to get at least one valid recipe quickly
	singleTree, singleTime, singleNodes := DFS(target, elements)
	
	// Track total nodes visited
	var totalNodesVisited int32 = int32(singleNodes)
	
	// Create a map to track unique recipes
	var resultsLock sync.Mutex
	var results []models.RecipeTree
	uniqueRecipes := make(map[string]bool)
	
	// If DFS found a recipe, use it as our first result
	if singleTree.Root != "" { // Check if valid tree
		resultsLock.Lock()
		treeKey := generateTreeKey(singleTree)
		uniqueRecipes[treeKey] = true
		results = append(results, singleTree)
		resultsLock.Unlock()
		fmt.Printf("Found first recipe using single DFS in %.2f seconds\n", singleTime)
	} else {
		fmt.Printf("Single DFS couldn't find a recipe for %s\n", target)
	}
	
	// If we only wanted one recipe or already have the max, return now
	if maxRecipes <= 1 || len(results) >= maxRecipes {
		return results, time.Since(startTime).Seconds(), int(totalNodesVisited)
	}
	
	// Setup for multithreaded recipe generation
	var wg sync.WaitGroup
	done := make(chan struct{}) // Channel to signal early completion
	
	// Create a channel to collect results from worker goroutines
	recipeChannel := make(chan models.RecipeTree, 100)
	
	// Start a goroutine to collect recipes from the channel
	go func() {
		for recipeTree := range recipeChannel {
			resultsLock.Lock()
			
			// Only add if we haven't reached max yet
			if len(results) < maxRecipes {
				treeKey := generateTreeKey(recipeTree)
				
				// Add if not a duplicate
				if !uniqueRecipes[treeKey] {
					uniqueRecipes[treeKey] = true
					results = append(results, recipeTree)
					fmt.Printf("Added recipe %d of %d to results\n", len(results), maxRecipes)
					
					// Signal done if we've reached max recipes
					if len(results) >= maxRecipes {
						close(done)
						break
					}
				}
			}
			
			resultsLock.Unlock()
		}
	}()
	
	// Process each top-level recipe in a separate goroutine
	for i, element := range targetElements {
		if len(element.Recipes) != 2 {
			continue
		}
		
		comp1 := element.Recipes[0]
		comp2 := element.Recipes[1]
		
		// Skip if components don't exist in element map
		if _, exists := elementMap[comp1]; !exists {
			continue
		}
		
		if _, exists := elementMap[comp2]; !exists {
			continue
		}
		
		// Launch a goroutine for each recipe
		wg.Add(1)
		go func(recipeIndex int, component1, component2 string) {
			defer wg.Done()
			
			// Create a timeout for this goroutine
			timeout := time.After(2 * time.Second)
			
			// Create done channel for early termination
			workerDone := make(chan struct{})
			
			// Process this recipe in yet another goroutine to handle timeout
			go func() {
				defer close(workerDone)
				
				// First attempt to create a tree with this recipe
				tree := createBasicRecipeTree(target, component1, component2, elementMap, &totalNodesVisited)
				
				// Check if we created a valid tree
				if tree.Root != "" {
					// Try to send to channel, non-blocking in case channel is full
					select {
					case recipeChannel <- tree:
						fmt.Printf("Worker %d: Added recipe for %s + %s\n", recipeIndex, component1, component2)
					case <-done:
						// Skip if we're already done
						return
					default:
						// Skip if channel is full
					}
				}
			}()
			
			// Wait for either completion, timeout, or done signal
			select {
			case <-workerDone:
				// Worker finished normally
			case <-timeout:
				fmt.Printf("Worker %d timed out\n", recipeIndex)
			case <-done:
				// We've reached max recipes, stop processing
				return
			}
			
		}(i, comp1, comp2)
	}
	
	// Wait for all goroutines or early termination
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()
	
	// Wait for workers to finish with timeout
	select {
	case <-waitCh:
		fmt.Println("All workers completed normally")
	case <-time.After(5 * time.Second):
		fmt.Println("Some workers timed out, proceeding with results collected so far")
	case <-done:
		fmt.Printf("Found all %d requested recipes, stopping other workers\n", maxRecipes)
	}
	
	// Close the recipe channel to signal collector to finish
	close(recipeChannel)
	
	// Give collector goroutine time to process remaining items
	time.Sleep(10 * time.Millisecond)
	
	fmt.Printf("MultipleDFS returning %d recipes for %s with %d nodes visited\n", 
		len(results), target, totalNodesVisited)
	
	return results, time.Since(startTime).Seconds(), int(totalNodesVisited)
}

// createBasicRecipeTree creates a recipe tree with non-empty children for non-basic elements
func createBasicRecipeTree(target, comp1, comp2 string, elementMap map[string][]models.Element, nodesVisited *int32) models.RecipeTree {
	// Process first component
	var comp1Tree models.RecipeTree
	var comp1TreeValid bool
	
	if IsBasicElement(comp1) {
		atomic.AddInt32(nodesVisited, 1)
		comp1Tree = models.RecipeTree{
			Root:     comp1,
			Left:     "",
			Right:    "",
			Tier:     "0",
			Children: []models.RecipeTree{},
		}
		comp1TreeValid = true
	} else {
		atomic.AddInt32(nodesVisited, 1)
		comp1Elem, exists := elementMap[comp1]
		// Non-basic element must have proper children
		if exists && len(comp1Elem) > 0 {
			// Find the first valid recipe for comp1
			for _, el := range comp1Elem {
				if len(el.Recipes) == 2 {
					subComp1 := el.Recipes[0]
					subComp2 := el.Recipes[1]
					
					// Create a proper tree for comp1
					comp1Tree = models.RecipeTree{
						Root:  comp1,
						Left:  subComp1,
						Right: subComp2,
						Tier:  fmt.Sprintf("%d", getTier(comp1, elementMap)),
						Children: []models.RecipeTree{
							{
								Root:     subComp1,
								Left:     "",
								Right:    "",
								Tier:     fmt.Sprintf("%d", getTier(subComp1, elementMap)),
								Children: []models.RecipeTree{},
							},
							{
								Root:     subComp2,
								Left:     "",
								Right:    "",
								Tier:     fmt.Sprintf("%d", getTier(subComp2, elementMap)),
								Children: []models.RecipeTree{},
							},
						},
					}
					atomic.AddInt32(nodesVisited, 2)
					comp1TreeValid = true
					break
				}
			}
		}
		
		// If we still don't have a proper tree for comp1
		if !comp1TreeValid {
			// Create a minimal node without children as fallback
			comp1Tree = models.RecipeTree{
				Root:     comp1,
				Left:     "",
				Right:    "",
				Tier:     fmt.Sprintf("%d", getTier(comp1, elementMap)),
				Children: []models.RecipeTree{},
			}
			comp1TreeValid = true
		}
	}
	
	// Process second component
	var comp2Tree models.RecipeTree
	var comp2TreeValid bool
	
	if IsBasicElement(comp2) {
		atomic.AddInt32(nodesVisited, 1)
		comp2Tree = models.RecipeTree{
			Root:     comp2,
			Left:     "",
			Right:    "",
			Tier:     "0",
			Children: []models.RecipeTree{},
		}
		comp2TreeValid = true
	} else {
		atomic.AddInt32(nodesVisited, 1)
		comp2Elem, exists := elementMap[comp2]
		// Non-basic element must have proper children
		if exists && len(comp2Elem) > 0 {
			// Find the first valid recipe for comp2
			for _, el := range comp2Elem {
				if len(el.Recipes) == 2 {
					subComp1 := el.Recipes[0]
					subComp2 := el.Recipes[1]
					
					// Create a proper tree for comp2
					comp2Tree = models.RecipeTree{
						Root:  comp2,
						Left:  subComp1,
						Right: subComp2,
						Tier:  fmt.Sprintf("%d", getTier(comp2, elementMap)),
						Children: []models.RecipeTree{
							{
								Root:     subComp1,
								Left:     "",
								Right:    "",
								Tier:     fmt.Sprintf("%d", getTier(subComp1, elementMap)),
								Children: []models.RecipeTree{},
							},
							{
								Root:     subComp2,
								Left:     "",
								Right:    "",
								Tier:     fmt.Sprintf("%d", getTier(subComp2, elementMap)),
								Children: []models.RecipeTree{},
							},
						},
					}
					atomic.AddInt32(nodesVisited, 2)
					comp2TreeValid = true
					break
				}
			}
		}
		
		// If we still don't have a proper tree for comp2
		if !comp2TreeValid {
			// Create a minimal node without children as fallback
			comp2Tree = models.RecipeTree{
				Root:     comp2,
				Left:     "",
				Right:    "",
				Tier:     fmt.Sprintf("%d", getTier(comp2, elementMap)),
				Children: []models.RecipeTree{},
			}
			comp2TreeValid = true
		}
	}
	
	// Create the final tree
	if !comp1TreeValid || !comp2TreeValid {
		return models.RecipeTree{} // Empty invalid tree
	}
	
	return models.RecipeTree{
		Root:     target,
		Left:     comp1,
		Right:    comp2,
		Tier:     fmt.Sprintf("%d", getTier(target, elementMap)),
		Children: []models.RecipeTree{comp1Tree, comp2Tree},
	}
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