package search

import (
	"fmt"
	"sync"
	"time"
	"Tubes2_alchendol/models"
)

// MultipleDFS searches for multiple recipes using DFS with multithreading
func MultipleDFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
	// Cap maximum recipes at 100
	if maxRecipes > 100 {
		maxRecipes = 100
	}
	
	startTime := time.Now()
	
	// Filter elements based on tier constraint
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
	
	// Create channels and wait group for multithreading
	resultChan := make(chan RecipeResult, maxRecipes)
	var wg sync.WaitGroup
	
	// Track unique recipes to avoid duplicates
	uniqueRecipes := make(map[string]bool)
	var recipesMutex sync.Mutex
	
	// Track total nodes visited across all threads
	var totalNodesVisited int
	var nodesMutex sync.Mutex
	
	// Counter for found recipes
	var foundRecipes int
	var foundMutex sync.Mutex
	
	// Find all possible combinations for the target element
	recipeCombinations := findAllRecipeCombinations(target, elementMap)
	
	// Process combinations in parallel
	for i, combination := range recipeCombinations {
		// Check if we've already found enough recipes
		foundMutex.Lock()
		if foundRecipes >= maxRecipes {
			foundMutex.Unlock()
			break
		}
		foundMutex.Unlock()
		
		// Skip if we've already tried this exact combination
		recipeKey := fmt.Sprintf("%s+%s", combination[0], combination[1])
		recipesMutex.Lock()
		if uniqueRecipes[recipeKey] {
			recipesMutex.Unlock()
			continue
		}
		uniqueRecipes[recipeKey] = true
		recipesMutex.Unlock()
		
		wg.Add(1)
		go func(comp1, comp2 string, index int) {
			defer wg.Done()
			
			// Find recipe variations for this combination using DFS approach
			findDFSRecipeVariations(comp1, comp2, target, elementMap, resultChan, 
				&foundRecipes, maxRecipes, &totalNodesVisited, &foundMutex, &nodesMutex)
			
		}(combination[0], combination[1], i)
	}
	
	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results
	var recipes []models.RecipeTree
	recipeTreeMap := make(map[string]bool) // To track unique recipe trees
	
	for result := range resultChan {
		if result.Found {
			// Generate a key to check for duplicate recipes
			recipeKey := generateTreeKey(result.Recipe)
			
			if !recipeTreeMap[recipeKey] {
				recipeTreeMap[recipeKey] = true
				recipes = append(recipes, result.Recipe)
				
				// Break if we have enough recipes
				if len(recipes) >= maxRecipes {
					break
				}
			}
		}
	}
	
	return recipes, time.Since(startTime).Seconds(), totalNodesVisited
}

// RecipeResult represents a result from concurrent search
type RecipeResult struct {
	Recipe models.RecipeTree
	Found  bool
}

// findDFSRecipeVariations finds recipe variations using DFS approach
func findDFSRecipeVariations(comp1, comp2, target string, elementMap map[string][]models.Element, 
	resultChan chan<- RecipeResult, foundRecipes *int, maxRecipes int, 
	totalNodesVisited *int, foundMutex, nodesMutex *sync.Mutex) {
	
	// Check if we've already found enough recipes
	foundMutex.Lock()
	if *foundRecipes >= maxRecipes {
		foundMutex.Unlock()
		return
	}
	foundMutex.Unlock()
	
	// Create a DFS state for each component
	state1 := &DFSState{
		Target:       comp1,
		ElementMap:   elementMap,
		Path:         make([]string, 0),
		NodesVisited: 0,
		Cache:        NewRecipeCache(),
	}
	
	state2 := &DFSState{
		Target:       comp2,
		ElementMap:   elementMap,
		Path:         make([]string, 0),
		NodesVisited: 0,
		Cache:        NewRecipeCache(),
	}
	
	// Find recipe for comp1 using DFS
	comp1Node, found1 := dfsSearchRecipe(state1, comp1)
	if !found1 {
		return
	}
	
	// Find recipe for comp2 using DFS
	comp2Node, found2 := dfsSearchRecipe(state2, comp2)
	if !found2 {
		return
	}
	
	// Update total nodes visited
	nodesMutex.Lock()
	*totalNodesVisited += state1.NodesVisited + state2.NodesVisited + 1
	nodesMutex.Unlock()
	
	// Check if we've already found enough recipes
	foundMutex.Lock()
	if *foundRecipes >= maxRecipes {
		foundMutex.Unlock()
		return
	}
	*foundRecipes++
	foundMutex.Unlock()
	
	// Create recipe tree for the target
	comp1Tree := ConvertToRecipeTree(comp1Node, elementMap)
	comp2Tree := ConvertToRecipeTree(comp2Node, elementMap)
	
	recipeTree := models.RecipeTree{
		Root:     target,
		Left:     comp1,
		Right:    comp2,
		Tier:     getTierAsString(target, elementMap),
		Children: []models.RecipeTree{comp1Tree, comp2Tree},
	}
	
	// Send result through channel
	resultChan <- RecipeResult{
		Recipe: recipeTree,
		Found:  true,
	}
}

// Helper function to find all possible recipe combinations for a target element
func findAllRecipeCombinations(target string, elementMap map[string][]models.Element) [][]string {
	var combinations [][]string
	
	// Get all recipes for the target element
	targetElements, found := elementMap[target]
	if !found {
		return combinations
	}
	
	// Track unique combinations
	uniqueCombinations := make(map[string]bool)
	
	for _, element := range targetElements {
		if len(element.Recipes) != 2 {
			continue
		}
		
		comp1 := element.Recipes[0]
		comp2 := element.Recipes[1]
		
		// Check tier constraints
		comp1Tier := getTier(comp1, elementMap)
		comp2Tier := getTier(comp2, elementMap)
		targetTier := getTier(target, elementMap)
		
		if comp1Tier >= targetTier || comp2Tier >= targetTier {
			continue
		}
		
		// Sort components for consistent key generation
		sortedComp1, sortedComp2 := comp1, comp2
		if sortedComp1 > sortedComp2 {
			sortedComp1, sortedComp2 = sortedComp2, sortedComp1
		}
		
		// Create a key for this combination
		key := fmt.Sprintf("%s+%s", sortedComp1, sortedComp2)
		if !uniqueCombinations[key] {
			uniqueCombinations[key] = true
			combinations = append(combinations, []string{comp1, comp2})
		}
	}
	
	return combinations
}

// Helper functions (need to be copied from common.go or imported)
func generateTreeKey(tree models.RecipeTree) string {
	if len(tree.Children) == 0 {
		return tree.Root
	}
	
	leftKey := ""
	rightKey := ""
	
	if len(tree.Children) > 0 {
		leftKey = generateTreeKey(tree.Children[0])
	}
	
	if len(tree.Children) > 1 {
		rightKey = generateTreeKey(tree.Children[1])
	}
	
	// Sort children keys for consistent ordering
	if leftKey > rightKey && rightKey != "" {
		return fmt.Sprintf("%s:(%s+%s)", tree.Root, rightKey, leftKey)
	} else {
		return fmt.Sprintf("%s:(%s+%s)", tree.Root, leftKey, rightKey)
	}
}

func getTier(element string, elementMap map[string][]models.Element) int {
	if IsBasicElement(element) {
		return 0
	}
	
	elements, found := elementMap[element]
	if !found || len(elements) == 0 {
		return -1 // Unknown element
	}
	
	return elements[0].Tier
}

func getTierAsString(element string, elementMap map[string][]models.Element) string {
	return fmt.Sprintf("%d", getTier(element, elementMap))
}