package search

import (
	"fmt"
	"strconv"
	"time"

	"Tubes2_alchendol/models"
)

// Cache structure to store already computed recipe trees
type RecipeCache struct {
	cache map[string]models.RecipeNode
}

// Create a new recipe cache
func NewRecipeCache() *RecipeCache {
	return &RecipeCache{
		cache: make(map[string]models.RecipeNode),
	}
}

// Get a recipe from cache
func (c *RecipeCache) Get(elementName string) (models.RecipeNode, bool) {
	node, exists := c.cache[elementName]
	return node, exists
}

// Store a recipe in cache
func (c *RecipeCache) Set(elementName string, node models.RecipeNode) {
	c.cache[elementName] = node
}

// Check if element is a basic element
func IsBasicElement(elementName string) bool {
	return elementName == "Air" || elementName == "Earth" || elementName == "Fire" || elementName == "Water"
}

// Add basic elements to the element map
func AddBasicElements(elementMap map[string][]models.Element) {
	basicElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, name := range basicElements {
		// Only add if not already present
		if _, exists := elementMap[name]; !exists {
			elementMap[name] = []models.Element{{
				Name:    name,
				Recipes: []string{},
				Tier:    0,
			}}
		}
	}
}

// Fungsi untuk membuat map elemen yang difilter sesuai dengan aturan tier
func CreateFilteredElementMap(elements []models.Element, target string) (map[string][]models.Element, bool) {
	targetFound := false
	elementMap := make(map[string][]models.Element)
	targetTier := -1
	
	// Find target element and its tier
	for _, el := range elements {
		if el.Name == target {
			targetFound = true
			if el.Tier > targetTier {
				targetTier = el.Tier
			}
			elementMap[el.Name] = append(elementMap[el.Name], el)
		}
	}
	
	if !targetFound {
		fmt.Printf("DEBUG: Target '%s' not found in elements\n", target)
		return elementMap, false
	}
	
	// Add elements with lower tier than target or basic elements
	for _, el := range elements {
		if el.Name == target {
			continue  // Already added
		}
		
		// Add element if its tier is lower than target's tier
		if el.Tier < targetTier {
			elementMap[el.Name] = append(elementMap[el.Name], el)
		}
	}
	
	// Make sure we add basic elements
	for _, el := range elements {
		if el.Tier == 0 {
			elementMap[el.Name] = append(elementMap[el.Name], el)
		}
	}
	
	return elementMap, true
}

// Struktur untuk menyimpan state pencarian DFS
type DFSState struct {
	Target       string
	ElementMap   map[string][]models.Element
	Path         []string
	NodesVisited int
	Cache        *RecipeCache
}

// Fungsi utama untuk DFS - mencari satu recipe terpendek
func DFS(target string, elements []models.Element) (models.RecipeTree, float64, int) {
	// Initialize start time
	startTime := time.Now()
	
	// Filter elements based on tier constraint
	elementMap, targetFound := CreateFilteredElementMap(elements, target)
	if !targetFound {
		return models.RecipeTree{}, time.Since(startTime).Seconds(), 0
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
		
		return basicTree, time.Since(startTime).Seconds(), 1
	}
	
	// Initialize DFS state with cache
	state := &DFSState{
		Target:       target,
		ElementMap:   elementMap,
		Path:         make([]string, 0),
		NodesVisited: 0,
		Cache:        NewRecipeCache(),
	}
	
	// Start DFS search with cache
	recipeNode, found := dfsSearchRecipe(state, target)
	if !found {
		return models.RecipeTree{}, time.Since(startTime).Seconds(), state.NodesVisited
	}
	
	// Convert the RecipeNode to models.RecipeTree
	recipeTree := ConvertToRecipeTree(recipeNode, elementMap)
	
	return recipeTree, time.Since(startTime).Seconds(), state.NodesVisited
}

// Helper function untuk memeriksa apakah elemen ada dalam path
func isInPath(path []string, element string) bool {
	for _, p := range path {
		if p == element {
			return true
		}
	}
	return false
}

// Get tier of element
func getElementTier(elementName string, elementMap map[string][]models.Element) int {
	if elementSlice, exists := elementMap[elementName]; exists && len(elementSlice) > 0 {
		return elementSlice[0].Tier
	}
	
	// Default to 0 for basic elements
	if IsBasicElement(elementName) {
		return 0
	}
	
	return -1
}

// Fungsi rekursif untuk mencari resep menggunakan DFS
func dfsSearchRecipe(state *DFSState, elementName string) (models.RecipeNode, bool) {
	// Count node visit
	state.NodesVisited++
	
	// Check cache first
	if cachedNode, exists := state.Cache.Get(elementName); exists {
		return cachedNode, true
	}
	
	// If basic element, return directly
	if IsBasicElement(elementName) {
		node := models.RecipeNode{
			Element: elementName,
			IsBasic: true,
		}
		state.Cache.Set(elementName, node)
		return node, true
	}
	
	// Check for cycles
	if isInPath(state.Path, elementName) {
		return models.RecipeNode{}, false
	}
	
	// Add to path
	state.Path = append(state.Path, elementName)
	defer func() {
		// Remove from path when done
		if len(state.Path) > 0 {
			state.Path = state.Path[:len(state.Path)-1]
		}
	}()
	
	// Get element from map
	elementSlice, found := state.ElementMap[elementName]
	if !found || len(elementSlice) == 0 {
		return models.RecipeNode{}, false
	}
	
	// Get the tier of the current element
	elementTier := getElementTier(elementName, state.ElementMap)
	
	// Try each element's recipes
	for _, element := range elementSlice {
		// Skip if recipes are not valid
		if len(element.Recipes) != 2 {
			continue
		}
		
		comp1 := element.Recipes[0]
		comp2 := element.Recipes[1]
		
		// Check tier constraints
		comp1Tier := getElementTier(comp1, state.ElementMap)
		comp2Tier := getElementTier(comp2, state.ElementMap)
		
		if comp1Tier == -1 || comp2Tier == -1 {
			continue
		}
		
		if comp1Tier >= elementTier || comp2Tier >= elementTier {
			continue
		}
		
		// Recursively find recipes for components
		comp1Node, found1 := dfsSearchRecipe(state, comp1)
		if !found1 {
			continue
		}
		
		comp2Node, found2 := dfsSearchRecipe(state, comp2)
		if !found2 {
			continue
		}
		
		// Create recipe node
		node := models.RecipeNode{
			Element:    elementName,
			Components: []models.RecipeNode{comp1Node, comp2Node},
			IsBasic:    false,
		}
		
		// Cache the result
		state.Cache.Set(elementName, node)
		
		return node, true
	}
	
	return models.RecipeNode{}, false
}

// Convert RecipeNode to models.RecipeTree format for JSON output
func ConvertToRecipeTree(node models.RecipeNode, elementMap map[string][]models.Element) models.RecipeTree {
	// Get the tier of the element
	tier := "0"
	if !node.IsBasic {
		if elementSlice, exists := elementMap[node.Element]; exists && len(elementSlice) > 0 {
			tier = strconv.Itoa(elementSlice[0].Tier)
		}
	}
	
	// Create the base models.RecipeTree
	recipeTree := models.RecipeTree{
		Root:     node.Element,
		Left:     "",
		Right:    "",
		Tier:     tier,
		Children: []models.RecipeTree{},
	}
	
	// If it's a basic element, we're done
	if node.IsBasic {
		return recipeTree
	}
	
	// For non-basic elements, add the components
	if len(node.Components) >= 2 {
		recipeTree.Left = node.Components[0].Element
		recipeTree.Right = node.Components[1].Element
		
		// Recursively convert the components
		leftTree := ConvertToRecipeTree(node.Components[0], elementMap)
		rightTree := ConvertToRecipeTree(node.Components[1], elementMap)
		
		// Add them to the children array
		recipeTree.Children = append(recipeTree.Children, leftTree, rightTree)
	}
	
	return recipeTree
}