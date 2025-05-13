package search

import (
	"Tubes2_alchendol/models"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"container/list"
)

func MultipleBFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	// Cap maximum recipes at a reasonable number
	if maxRecipes > 100 {
		maxRecipes = 100
	}
	
	startTime := time.Now()
	
	// Determine timeout based on element complexity
	timeoutDuration := 15 * time.Second // Default timeout
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
	
	// Track results and nodes visited
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
	
	// Start a goroutine to collect node visit counts
	go func() {
		for count := range visitedChannel {
			atomic.AddInt32(&totalNodesVisited, int32(count))
		}
	}()
	
	// Start a collector goroutine to process the trees
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
			treeKey := generateBFSTreeKey(tree)
			
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
			
			// Create a recipe tree using BFS for this combination
			tree := bfsCreateRecipeTree(target, c1, c2, elementMap, visitedChannel)
			
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
			bfsExploreAlternativeTrees(target, c1, c2, elementMap, treeChannel, visitedChannel, ctx, done)
			
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
	
	fmt.Printf("MultipleBFS returning %d recipes for %s with %d nodes visited\n", 
		len(results), target, finalNodesVisited)
	
	return results, time.Since(startTime).Seconds(), int(finalNodesVisited)
}

// BFS for creating a recipe tree from a given combination
func bfsCreateRecipeTree(target, comp1, comp2 string, elementMap map[string][]models.Element, visitedChannel chan<- int) models.RecipeTree {
	// Increment visit count
	visitedChannel <- 1
	
	// Check tier constraints
	targetTier := getElementTier(target, elementMap)
	comp1Tier := getElementTier(comp1, elementMap)
	comp2Tier := getElementTier(comp2, elementMap)
	
	if comp1Tier >= targetTier || comp2Tier >= targetTier {
		return models.RecipeTree{}
	}
	
	// Create root node for the target
	root := models.RecipeTree{
		Root:     target,
		Left:     comp1,
		Right:    comp2,
		Tier:     fmt.Sprintf("%d", targetTier),
		Children: []models.RecipeTree{},
	}
	
	// Perform BFS for each component
	comp1Tree := bfsSearchComponent(comp1, elementMap, visitedChannel)
	comp2Tree := bfsSearchComponent(comp2, elementMap, visitedChannel)
	
	// If either component search failed, return empty tree
	if comp1Tree.Root == "" || comp2Tree.Root == "" {
		return models.RecipeTree{}
	}
	
	// Add component trees as children
	root.Children = append(root.Children, comp1Tree, comp2Tree)
	
	return root
}

// BFS search for a component recipe path
func bfsSearchComponent(element string, elementMap map[string][]models.Element, visitedChannel chan<- int) models.RecipeTree {
	// Increment visit count
	visitedChannel <- 1
	
	// If element is basic, return directly
	if IsBasicElement(element) {
		return models.RecipeTree{
			Root:     element,
			Left:     "",
			Right:    "",
			Tier:     "0",
			Children: []models.RecipeTree{},
		}
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
		visitedChannel <- 1
		
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
	return buildTreeFromNodeMap(element, nodeMap)
}

// Build a complete tree from the node map
func buildTreeFromNodeMap(root string, nodeMap map[string]models.RecipeTree) models.RecipeTree {
	result := nodeMap[root]
	
	// If this is a basic element or has no children, return as is
	if result.Left == "" || result.Right == "" {
		return result
	}
	
	// Recursively build children
	leftChild := buildTreeFromNodeMap(result.Left, nodeMap)
	rightChild := buildTreeFromNodeMap(result.Right, nodeMap)
	
	// Add children to result
	result.Children = []models.RecipeTree{leftChild, rightChild}
	
	return result
}

// Explore alternative trees using BFS
func bfsExploreAlternativeTrees(target, comp1, comp2 string, elementMap map[string][]models.Element, 
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
	targetTier := getElementTier(target, elementMap)
	
	// Skip basic elements
	if IsBasicElement(comp1) && IsBasicElement(comp2) {
		return
	}
	
	// Find all alternative recipes for the first component
	if !IsBasicElement(comp1) && comp1 != target { // Avoid cycles
		bfsExploreComponentCombinations(comp1, elementMap, visitedChannel, func(altComp1Tree models.RecipeTree) {
			// Find a standard path for comp2 using BFS
			comp2Tree := bfsSearchComponent(comp2, elementMap, visitedChannel)
			
			if comp2Tree.Root == "" {
				return
			}
			
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
		bfsExploreComponentCombinations(comp2, elementMap, visitedChannel, func(altComp2Tree models.RecipeTree) {
			// Find a standard path for comp1 using BFS
			comp1Tree := bfsSearchComponent(comp1, elementMap, visitedChannel)
			
			if comp1Tree.Root == "" {
				return
			}
			
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
		bfsExploreComponentCombinations(comp1, elementMap, visitedChannel, func(altComp1Tree models.RecipeTree) {
			bfsExploreComponentCombinations(comp2, elementMap, visitedChannel, func(altComp2Tree models.RecipeTree) {
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

// Explore component combinations using BFS
func bfsExploreComponentCombinations(component string, elementMap map[string][]models.Element, 
	visitedChannel chan<- int, callback func(models.RecipeTree)) {
	
	// Count this function call as a node visit
	visitedChannel <- 1
	
	// Get component tier
	compTier := getElementTier(component, elementMap)
	
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
		subComp1Tier := getElementTier(subComp1, elementMap)
		subComp2Tier := getElementTier(subComp2, elementMap)
		
		// Skip tier violations
		if subComp1Tier >= compTier || subComp2Tier >= compTier {
			continue
		}
		
		// Create tree for this combination using BFS
		tree := bfsCreateRecipeTree(component, subComp1, subComp2, elementMap, visitedChannel)
		
		// Call the callback with this tree if valid
		if tree.Root != "" {
			callback(tree)
		}
	}
}

// Helper function to generate a BFS-style key for a tree
func generateBFSTreeKey(tree models.RecipeTree) string {
	var sb strings.Builder
	generateBFSTreeKeyHelper(tree, &sb, 0)
	return sb.String()
}

func generateBFSTreeKeyHelper(tree models.RecipeTree, sb *strings.Builder, depth int) {
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
		
		generateBFSTreeKeyHelper(left, sb, depth+1)
		sb.WriteString("+")
		generateBFSTreeKeyHelper(right, sb, depth+1)
	} else if len(tree.Children) == 1 {
		generateBFSTreeKeyHelper(tree.Children[0], sb, depth+1)
	}
	
	sb.WriteString(")")
}
