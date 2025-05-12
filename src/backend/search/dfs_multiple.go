package search

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"os"
	"context"
	"Tubes2_alchendol/models"
)

// MultipleDFS searches for multiple recipes using concurrent DFS with multithreading
func MultipleDFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	// Cap maximum recipes at a reasonable number
	if maxRecipes > 100 {
		maxRecipes = 100
	}
	
	startTime := time.Now()
	
	// Determine timeout based on element complexity
	timeoutDuration := 15 * time.Second // Default from your original code
	for _, el := range elements {
		if el.Name == target {
			// Complex elements get more time based on tier
			if el.Tier > 5 {
				timeoutDuration = 60 * time.Second
			} else if el.Tier > 3 {
				timeoutDuration = 30 * time.Second
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
	
	// Track results and nodes visited - KEEPING EXACTLY AS IN ORIGINAL
	var results []models.RecipeTree
	var uniqueRecipes = make(map[string]bool)
	var totalNodesVisited int32 = 0 // Akan diinkremen di setiap langkah pencarian
	var resultsMutex sync.Mutex
	
	// Create atomic counter for unique recipes found
	var uniqueRecipesCounter int32 = 0
	
	// Get all direct recipes for the target
	targetElements, found := elementMap[target]
	if !found || len(targetElements) == 0 {
		fmt.Printf("No recipes found for target '%s' in elementMap\n", target)
		return []models.RecipeTree{}, time.Since(startTime).Seconds(), 0
	}
	
	// Create a channel for recipe trees from workers
	treeChannel := make(chan models.RecipeTree, 100)
	
	// Create a channel for reporting nodes visited
	visitedChannel := make(chan int, 100)
	
	// Signal channel for early termination
	done := make(chan struct{})
	
	// Worker counter
	var waitGroup sync.WaitGroup
	
	// Start a goroutine to collect node visit counts - EXACTLY AS ORIGINAL
	go func() {
		for count := range visitedChannel {
			atomic.AddInt32(&totalNodesVisited, int32(count))
		}
	}()
	
	// Start a collector goroutine to process the trees - EXACTLY AS ORIGINAL
	go func() {
		for tree := range treeChannel {
			if tree.Root == "" {
				continue
			}
			
			resultsMutex.Lock()
			
			// Check if we've reached max recipes
			if len(results) >= maxRecipes {
				resultsMutex.Unlock()
				continue
			}
			
			// Generate a unique key for this tree
			treeKey := generateCompleteTreeKey(tree)
			
			// Only add if we haven't seen this tree before
			if !uniqueRecipes[treeKey] {
				uniqueRecipes[treeKey] = true
				results = append(results, tree)
				
				// Increment our atomic counter for unique recipes
				newCount := atomic.AddInt32(&uniqueRecipesCounter, 1)
				
				// Tampilkan juga jumlah node yang dikunjungi
				nodeCount := atomic.LoadInt32(&totalNodesVisited)
				fmt.Printf("Added recipe %d/%d: %s = %s + %s (Unique tree: %d, NodesVisited: %d)\n", 
					len(results), maxRecipes, tree.Root, tree.Left, tree.Right, newCount, nodeCount)
				
				// If we've reached max recipes, signal done
				if len(results) >= maxRecipes {
					select {
					case done <- struct{}{}:
					default:
					}
				}
			}
			
			resultsMutex.Unlock()
		}
	}()

	// First, process all direct recipe combinations
	processedCombos := make(map[string]bool)
	var directCombinations []struct {
		Comp1, Comp2 string
	}
	
	// Collect unique direct combinations
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
		directCombinations = append(directCombinations, struct {
			Comp1, Comp2 string
		}{
			Comp1: comp1,
			Comp2: comp2,
		})
	}
	
	fmt.Printf("Found %d unique direct combinations for %s\n", len(directCombinations), target)
	
	// Process each direct combination
	for i, combo := range directCombinations {
		// Check for early termination or timeout
		select {
		case <-done:
			fmt.Printf("Found all %d requested recipes, skipping additional combinations\n", maxRecipes)
			goto ProcessComplete
		case <-ctx.Done():
			fmt.Println("Context timeout reached, using results gathered so far")
			goto ProcessComplete
		default:
			// Continue processing
		}
		
		waitGroup.Add(1)
		
		go func(index int, c1, c2 string) {
			defer waitGroup.Done()
			
			// Create a new state for this combination
			state := &DFSState{
				Target:       target,
				ElementMap:   elementMap,
				Path:         make([]string, 0),
				NodesVisited: 0,
				Cache:        NewRecipeCache(),
			}
			
			// First, try the direct combination
			tree := createCompleteRecipeTree(target, c1, c2, state)
			
			// Report nodes visited
			visitedChannel <- state.NodesVisited
			
			// Send the tree to the channel if valid
			if tree.Root != "" {
				select {
				case treeChannel <- tree:
				case <-ctx.Done():
					return
				case <-done:
					return
				}
			}
			
			// Now explore alternative trees for this combination's components
			exploreAlternativeTrees(target, c1, c2, elementMap, treeChannel, visitedChannel, ctx, done)
			
		}(i, combo.Comp1, combo.Comp2)
	}
	
	// Wait for all workers to finish or timeout
	go func() {
		waitGroup.Wait()
		close(treeChannel)
		close(visitedChannel)
	}()
	
	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		fmt.Println("Processing timed out, using results gathered so far")
	case <-done:
		fmt.Printf("Found all %d requested recipes\n", maxRecipes)
	}
	
ProcessComplete:
	// Wait up to 5 seconds for workers to complete, then return what we've got
	cleanupTimeout := time.After(5 * time.Second)
	select {
	case <-func() chan struct{} {
		done := make(chan struct{})
		go func() {
			waitGroup.Wait()
			close(done)
		}()
		return done
	}():
		fmt.Println("All workers completed normally")
	case <-cleanupTimeout:
		fmt.Println("Some workers still running, proceeding with results collected so far")
	case <-ctx.Done():
		fmt.Println("Main context timeout, returning available results")
	}
	
	// Debug output if enabled
	if os.Getenv("DEBUG_RECIPES") == "1" && len(results) > 1 {
		DebugDisplayRecipeDifferences(results)
	}
	
	// Get the final count of unique recipes found
	finalUniqueCount := atomic.LoadInt32(&uniqueRecipesCounter)
	finalNodesVisited := atomic.LoadInt32(&totalNodesVisited)
	
	fmt.Printf("Total unique recipe combinations found: %d\n", finalUniqueCount)
	fmt.Printf("Total nodes visited during search: %d\n", finalNodesVisited)
	
	fmt.Printf("MultipleDFS returning %d recipes for %s with %d nodes visited\n", 
		len(results), target, finalNodesVisited)
	
	return results, time.Since(startTime).Seconds(), int(finalNodesVisited)
}

// Modifikasi function exploreAlternativeTrees untuk menerima visitedChannel - KEEPING AS ORIGINAL
func exploreAlternativeTrees(target, comp1, comp2 string, elementMap map[string][]models.Element, 
	treeChannel chan<- models.RecipeTree, visitedChannel chan<- int, ctx context.Context, done <-chan struct{}) {
	
	// Check for context cancellation or done signal
	select {
	case <-ctx.Done():
		return
	case <-done:
		return
	default:
		// Continue
	}
	
	// Count this function call as a node visit
	visitedChannel <- 1
	
	// Get target tier
	targetTier := -1
	if elements, found := elementMap[target]; found && len(elements) > 0 {
		targetTier = elements[0].Tier
	}
	
	// Skip basic elements
	if IsBasicElement(comp1) && IsBasicElement(comp2) {
		return
	}
	
	// Find all alternative recipes for the first component
	if !IsBasicElement(comp1) && comp1 != target { // Avoid cycles
		exploreComponentCombinations(comp1, elementMap, visitedChannel, func(altComp1Tree models.RecipeTree) {
			// For each alternative for comp1, create a standard tree for comp2
			state := &DFSState{
				Target:       comp2,
				ElementMap:   elementMap,
				Path:         make([]string, 0),
				NodesVisited: 0,
				Cache:        NewRecipeCache(),
			}
			
			// Find a standard path for comp2
			comp2Node, found := dfsSearchRecipe(state, comp2)
			
			// Report nodes visited
			visitedChannel <- state.NodesVisited
			
			if !found {
				return
			}
			
			// Convert to RecipeTree
			comp2Tree := ConvertToRecipeTree(comp2Node, elementMap)
			
			// Create the full tree with this alternative
			fullTree := models.RecipeTree{
				Root:  target,
				Left:  comp1,
				Right: comp2,
				Tier:  fmt.Sprintf("%d", targetTier),
				Children: []models.RecipeTree{
					altComp1Tree,
					comp2Tree,
				},
			}
			
			// Send to the channel
			select {
			case treeChannel <- fullTree:
			case <-ctx.Done():
				return
			case <-done:
				return
			default:
				// Channel full, skip
			}
		})
	}
	
	// Find all alternative recipes for the second component
	if !IsBasicElement(comp2) && comp2 != target { // Avoid cycles
		exploreComponentCombinations(comp2, elementMap, visitedChannel, func(altComp2Tree models.RecipeTree) {
			// For each alternative for comp2, create a standard tree for comp1
			state := &DFSState{
				Target:       comp1,
				ElementMap:   elementMap,
				Path:         make([]string, 0),
				NodesVisited: 0,
				Cache:        NewRecipeCache(),
			}
			
			// Find a standard path for comp1
			comp1Node, found := dfsSearchRecipe(state, comp1)
			
			// Report nodes visited
			visitedChannel <- state.NodesVisited
			
			if !found {
				return
			}
			
			// Convert to RecipeTree
			comp1Tree := ConvertToRecipeTree(comp1Node, elementMap)
			
			// Create the full tree with this alternative
			fullTree := models.RecipeTree{
				Root:  target,
				Left:  comp1,
				Right: comp2,
				Tier:  fmt.Sprintf("%d", targetTier),
				Children: []models.RecipeTree{
					comp1Tree,
					altComp2Tree,
				},
			}
			
			// Send to the channel
			select {
			case treeChannel <- fullTree:
			case <-ctx.Done():
				return
			case <-done:
				return
			default:
				// Channel full, skip
			}
		})
	}
	
	// Now try exploring even deeper combinations - variations of both components
	if !IsBasicElement(comp1) && !IsBasicElement(comp2) && comp1 != target && comp2 != target {
		// Count exploring deeper combinations as a node visit
		visitedChannel <- 1
		
		// Find different combinations for both components
		exploreComponentCombinations(comp1, elementMap, visitedChannel, func(altComp1Tree models.RecipeTree) {
			exploreComponentCombinations(comp2, elementMap, visitedChannel, func(altComp2Tree models.RecipeTree) {
				// Create a full tree with both alternative components
				fullTree := models.RecipeTree{
					Root:  target,
					Left:  comp1,
					Right: comp2,
					Tier:  fmt.Sprintf("%d", targetTier),
					Children: []models.RecipeTree{
						altComp1Tree,
						altComp2Tree,
					},
				}
				
				// Send to the channel
				select {
				case treeChannel <- fullTree:
				case <-ctx.Done():
					return
				case <-done:
					return
				default:
					// Channel full, skip
				}
			})
		})
	}
}

// Modifikasi function exploreComponentCombinations untuk menerima visitedChannel - KEEPING AS ORIGINAL
func exploreComponentCombinations(component string, elementMap map[string][]models.Element, 
	visitedChannel chan<- int, callback func(models.RecipeTree)) {
	
	// Count this function call as a node visit
	visitedChannel <- 1
	
	// Get component tier
	compTier := -1
	if elements, found := elementMap[component]; found && len(elements) > 0 {
		compTier = elements[0].Tier
	}
	
	// Get all recipes for this component
	componentElements, found := elementMap[component]
	if !found || len(componentElements) == 0 {
		return
	}
	
	// Try each recipe
	processedCombos := make(map[string]bool)
	
	for _, element := range componentElements {
		// Count each recipe iteration as a node visit
		visitedChannel <- 1
		
		if len(element.Recipes) != 2 {
			continue
		}
		
		subComp1 := element.Recipes[0]
		subComp2 := element.Recipes[1]
		
		// Skip if components don't exist
		if _, exists := elementMap[subComp1]; !exists {
			continue
		}
		if _, exists := elementMap[subComp2]; !exists {
			continue
		}
		
		// Create a unique key for this combination
		var comboKey string
		if subComp1 < subComp2 {
			comboKey = subComp1 + "+" + subComp2
		} else {
			comboKey = subComp2 + "+" + subComp1
		}
		
		// Skip if we've processed this combination
		if processedCombos[comboKey] {
			continue
		}
		
		processedCombos[comboKey] = true
		
		// Check tier constraints
		subComp1Tier := -1
		if elements, found := elementMap[subComp1]; found && len(elements) > 0 {
			subComp1Tier = elements[0].Tier
		} else if IsBasicElement(subComp1) {
			subComp1Tier = 0
		}
		
		subComp2Tier := -1
		if elements, found := elementMap[subComp2]; found && len(elements) > 0 {
			subComp2Tier = elements[0].Tier
		} else if IsBasicElement(subComp2) {
			subComp2Tier = 0
		}
		
		// Skip tier violations
		if subComp1Tier >= compTier || subComp2Tier >= compTier {
			continue
		}
		
		// Create a new state for this component
		state := &DFSState{
			Target:       component,
			ElementMap:   elementMap,
			Path:         make([]string, 0),
			NodesVisited: 0,
			Cache:        NewRecipeCache(),
		}
		
		// Create tree for this combination
		tree := createCompleteRecipeTree(component, subComp1, subComp2, state)
		
		// Report nodes visited
		visitedChannel <- state.NodesVisited
		
		// Call the callback with this tree if valid
		if tree.Root != "" {
			callback(tree)
		}
	}
}

// Helper function to generate a more detailed key for a tree - KEEPING AS ORIGINAL
func generateCompleteTreeKey(tree models.RecipeTree) string {
	var sb strings.Builder
	generateCompleteTreeKeyHelper(tree, &sb, 0)
	return sb.String()
}

func generateCompleteTreeKeyHelper(tree models.RecipeTree, sb *strings.Builder, depth int) {
	// Limit recursion depth
	if depth > 10 {
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
		
		generateCompleteTreeKeyHelper(left, sb, depth+1)
		sb.WriteString("+")
		generateCompleteTreeKeyHelper(right, sb, depth+1)
	} else if len(tree.Children) == 1 {
		generateCompleteTreeKeyHelper(tree.Children[0], sb, depth+1)
	}
	
	sb.WriteString(")")
}


// Create a complete recipe tree with paths to basic elements
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


// ------------------------------------------------------------
// ---// Debugging and analysis functions
// Debug function to compare and display recipe differences
func DebugDisplayRecipeDifferences(recipes []models.RecipeTree) {
	if len(recipes) <= 1 {
		fmt.Println("Need at least 2 recipes to compare differences")
		return
	}

	// fmt.Println("\n=== COMPLETE RECIPE DIFFERENCES ANALYSIS ===")
	// fmt.Printf("Analyzing %d different recipes for %s\n", len(recipes), recipes[0].Root)

	// Compare each recipe pair
	for i := 0; i < len(recipes); i++ {
		for j := i + 1; j < len(recipes); j++ {
			// fmt.Printf("\n=============================================")
			// fmt.Printf("\nCOMPARING RECIPE %d AND RECIPE %d:\n", i+1, j+1)
			// fmt.Printf("=============================================\n")
			
			// First print a summary of each recipe
			// fmt.Printf("\nRecipe %d structure:\n", i+1)
			// printRecipeStructure(recipes[i], 0)
			
			// fmt.Printf("\nRecipe %d structure:\n", j+1)
			// printRecipeStructure(recipes[j], 0)
			
			// // Then show the detailed differences
			// fmt.Printf("\nDetailed differences:\n")
			diffCount := compareRecipeTreesDeep(recipes[i], recipes[j], 0)
			
			if diffCount == 0 {
				fmt.Printf("No differences found! These recipes are identical in structure.\n")
			} else {
				fmt.Printf("\nTotal of %d differences found.\n", diffCount)
			}
		}
	}
}

// Improved comparison function that checks the entire tree structure
func compareRecipeTreesDeep(tree1, tree2 models.RecipeTree, depth int) int {
	diffCount := 0
	// indent := strings.Repeat("  ", depth)
	
	// Compare root elements
	if tree1.Root != tree2.Root {
		// fmt.Printf("%s- Different elements: [%s] vs [%s]\n", indent, tree1.Root, tree2.Root)
		return 1 // Early return, trees are completely different
	}
	
	// Check if one has children and the other doesn't
	if (len(tree1.Children) == 0) != (len(tree2.Children) == 0) {
		// fmt.Printf("%s- [%s]: One recipe has children, the other doesn't\n", indent, tree1.Root)
		return 1
	}
	
	// If both are basic elements, no differences
	if len(tree1.Children) == 0 {
		return 0
	}
	
	// Check direct components
	if tree1.Left != tree2.Left || tree1.Right != tree2.Right {
		// fmt.Printf("%s- [%s]: Different components: [%s + %s] vs [%s + %s]\n", 
		// 	indent, tree1.Root, tree1.Left, tree1.Right, tree2.Left, tree2.Right)
		diffCount++
	}
	
	// Recursively compare children
	if len(tree1.Children) >= 2 && len(tree2.Children) >= 2 {
		// Try to match children in the most logical way
		var leftTree1, rightTree1, leftTree2, rightTree2 models.RecipeTree
		
		// Determine which children to compare with each other
		if tree1.Left == tree2.Left && tree1.Right == tree2.Right {
			// Direct match of components
			leftTree1 = tree1.Children[0]
			rightTree1 = tree1.Children[1]
			leftTree2 = tree2.Children[0]
			rightTree2 = tree2.Children[1]
		} else if tree1.Left == tree2.Right && tree1.Right == tree2.Left {
			// Components are swapped
			leftTree1 = tree1.Children[0]
			rightTree1 = tree1.Children[1]
			leftTree2 = tree2.Children[1]
			rightTree2 = tree2.Children[0]
		} else {
			// Components don't match directly, try to match by name
			if tree1.Children[0].Root == tree2.Children[0].Root {
				leftTree1 = tree1.Children[0]
				leftTree2 = tree2.Children[0]
			} else if tree1.Children[0].Root == tree2.Children[1].Root {
				leftTree1 = tree1.Children[0]
				leftTree2 = tree2.Children[1]
			} else {
				leftTree1 = tree1.Children[0]
				leftTree2 = tree2.Children[0] // Best guess
			}
			
			if tree1.Children[1].Root == tree2.Children[1].Root {
				rightTree1 = tree1.Children[1]
				rightTree2 = tree2.Children[1]
			} else if tree1.Children[1].Root == tree2.Children[0].Root {
				rightTree1 = tree1.Children[1]
				rightTree2 = tree2.Children[0]
			} else {
				rightTree1 = tree1.Children[1]
				rightTree2 = tree2.Children[1] // Best guess
			}
		}
		
		// Compare matched children
		diffCount += compareRecipeTreesDeep(leftTree1, leftTree2, depth+1)
		diffCount += compareRecipeTreesDeep(rightTree1, rightTree2, depth+1)
	}
	
	return diffCount
}

// Improved structure printing that shows the complete tree
func printRecipeStructure(tree models.RecipeTree, depth int) {
	// indent := strings.Repeat("  ", depth)
	
	// if len(tree.Children) == 0 {
	// 	// Basic element
	// 	fmt.Printf("%s• %s (basic element)\n", indent, tree.Root)
	// 	return
	// }
	
	// // Print this element and its components
	// fmt.Printf("%s• %s = %s + %s\n", indent, tree.Root, tree.Left, tree.Right)
	
	// // Recursively print children
	// if len(tree.Children) >= 2 {
	// 	printRecipeStructure(tree.Children[0], depth+1)
	// 	printRecipeStructure(tree.Children[1], depth+1)
	// }
}