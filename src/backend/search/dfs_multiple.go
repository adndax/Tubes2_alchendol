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
	
	// Create a map to track unique recipes
	var resultsLock sync.Mutex
	var results []models.RecipeTree
	uniqueRecipes := make(map[string]bool)
	
	// Track total nodes visited
	var totalNodesVisited int32 = 0
	
	// Use DFS for the first recipe (reliable and complete to basic elements)
	fmt.Printf("Getting first recipe using DFS\n")
	singleTree, singleTime, singleNodes := DFS(target, elements)
	atomic.AddInt32(&totalNodesVisited, int32(singleNodes))
	
	// If DFS found a recipe, use it as our first result
	if singleTree.Root != "" {
		resultsLock.Lock()
		treeKey := generateTreeKey(singleTree)
		uniqueRecipes[treeKey] = true
		results = append(results, singleTree)
		resultsLock.Unlock()
		fmt.Printf("Found first recipe using DFS in %.2f seconds\n", singleTime)
	} else {
		fmt.Printf("DFS couldn't find a recipe for %s\n", target)
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
					}
				}
			}
			
			resultsLock.Unlock()
		}
	}()
	
	// Process each top-level recipe in a separate goroutine
	for i, element := range targetElements {
		// Skip if we've already found enough recipes
		select {
		case <-done:
			break
		default:
			// Continue processing
		}
		
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
			timeout := time.After(3 * time.Second)
			
			// Create done channel for early termination
			workerDone := make(chan struct{})
			
			// Process this recipe in yet another goroutine to handle timeout
			go func() {
				defer close(workerDone)
				
				// Create a state for tracking visited nodes
				state := &DFSState{
					Target:       target,
					ElementMap:   elementMap,
					Path:         make([]string, 0),
					NodesVisited: 0,
					Cache:        NewRecipeCache(),
				}
				
				// Create a complete recipe tree down to basic elements
				tree := createCompleteRecipeTree(target, component1, component2, state)
				
				// Check if we created a valid tree and send it
				if tree.Root != "" {
					// Update visited count atomically
					atomic.AddInt32(&totalNodesVisited, int32(state.NodesVisited))
					
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

// Create a recipe tree with complete paths to basic elements
func createCompleteRecipeTree(target, comp1, comp2 string, state *DFSState) models.RecipeTree {
	// Increment visit count
	state.NodesVisited++
	
	// Check tier constraints
	targetTier := getElementTier(target, state.ElementMap)
	comp1Tier := getElementTier(comp1, state.ElementMap)
	comp2Tier := getElementTier(comp2, state.ElementMap)
	
	if comp1Tier >= targetTier || comp2Tier >= targetTier {
		return models.RecipeTree{}
	}
	
	// Create component trees recursively to ensure paths to basic elements
	var comp1Tree, comp2Tree models.RecipeNode
	comp1Found, comp2Found := false, false
	
	// Process first component - try to get a complete path
	if IsBasicElement(comp1) {
		// Basic element is already a leaf
		comp1Tree = models.RecipeNode{
			Element: comp1,
			IsBasic: true,
		}
		comp1Found = true
	} else {
		// Try to find a valid recipe path for comp1
		comp1Node, found := dfsSearchRecipe(state, comp1)
		if found {
			comp1Tree = comp1Node
			comp1Found = true
		}
	}
	
	// Process second component - try to get a complete path
	if IsBasicElement(comp2) {
		// Basic element is already a leaf
		comp2Tree = models.RecipeNode{
			Element: comp2,
			IsBasic: true,
		}
		comp2Found = true
	} else {
		// Try to find a valid recipe path for comp2
		comp2Node, found := dfsSearchRecipe(state, comp2)
		if found {
			comp2Tree = comp2Node
			comp2Found = true
		}
	}
	
	// If either component failed, return empty tree
	if !comp1Found || !comp2Found {
		return models.RecipeTree{}
	}
	
	// Convert the RecipeNodes to RecipeTrees
	finalTree := models.RecipeTree{
		Root:  target,
		Left:  comp1,
		Right: comp2,
		Tier:  fmt.Sprintf("%d", targetTier),
		Children: []models.RecipeTree{
			ConvertToRecipeTree(comp1Tree, state.ElementMap),
			ConvertToRecipeTree(comp2Tree, state.ElementMap),
		},
	}
	
	return finalTree
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