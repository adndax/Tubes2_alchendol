package search

import (
	"Tubes2_alchendol/models"
	"container/list"
	"fmt"
	"time"
)

func BidirectionalSearch(targetElement string, elements []models.Element) (models.RecipeTree, float64, int) {
	startTime := time.Now()
	nodesVisited := 0

	elementMap := make(map[string]models.Element)
	for _, e := range elements {
		elementMap[e.Name] = e
	}

	recipeMap := createRecipeMap(elements)

	if _, exists := elementMap[targetElement]; !exists {
		fmt.Printf("Element '%s' not found in element database\n", targetElement)
		return models.RecipeTree{}, time.Since(startTime).Seconds(), nodesVisited
	}

	basicElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, basic := range basicElements {
		if basic == targetElement {
			return models.RecipeTree{
				Root:     targetElement,
				Left:     "",
				Right:    "",
				Tier:     "0",
				Children: []models.RecipeTree{},
			}, time.Since(startTime).Seconds(), nodesVisited
		}
	}

	forwardVisited := make(map[string]bool)
	backwardVisited := make(map[string]bool)
	
	forwardPath := make(map[string][]string)
	backwardPath := make(map[string][]string)
	
	bestTierPath := make(map[string]int)

	forwardQueue := list.New()
	
	for _, basic := range basicElements {
		forwardQueue.PushBack(&models.Node{
			Element: basic,
			Path:    []string{},
			Tier:    0, 
		})
		forwardVisited[basic] = true
		forwardPath[basic] = []string{}
		bestTierPath[basic] = 0
	}

	backwardQueue := list.New()
	targetTier := elementMap[targetElement].Tier
	backwardQueue.PushBack(&models.Node{
		Element: targetElement,
		Path:    []string{},
		Tier:    targetTier,
	})
	backwardVisited[targetElement] = true
	backwardPath[targetElement] = []string{}
	bestTierPath[targetElement] = targetTier

	type MeetingPoint struct {
		Element string
		Forward bool
		Path    []string
		Tier    int
	}
	
	var meetingPoints []MeetingPoint

	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 {
		if forwardQueue.Len() > 0 {
			current := forwardQueue.Remove(forwardQueue.Front()).(*models.Node)
			nodesVisited++

			for other := range forwardVisited {
				for result, recipes := range recipeMap {
					for _, recipe := range recipes {
						if (recipe[0] == current.Element && recipe[1] == other) ||
						   (recipe[1] == current.Element && recipe[0] == other) {
							// Found a new element we can create
							resultElement, exists := elementMap[result]
							if !exists {
								continue
							}
							
							resultTier := resultElement.Tier
							
							if !forwardVisited[result] || resultTier < bestTierPath[result] {
								// Only add if not visited or if we found a better tier path
								forwardQueue.PushBack(&models.Node{
									Element: result,
									Path:    append(current.Path, current.Element, other),
									Tier:    resultTier,
								})
								forwardVisited[result] = true
								forwardPath[result] = []string{current.Element, other}
								bestTierPath[result] = resultTier
								
								// Check if we've reached an element discovered by backward search
								if backwardVisited[result] {
									meetingPoints = append(meetingPoints, MeetingPoint{
										Element: result,
										Forward: true,
										Path:    append(current.Path, current.Element, other),
										Tier:    resultTier,
									})
								}
							}
						}
					}
				}
			}
		}

		// Backward BFS step (from target toward basic elements)
		if backwardQueue.Len() > 0 {
			current := backwardQueue.Remove(backwardQueue.Front()).(*models.Node)
			nodesVisited++

			// Get all direct recipes that can make the current element
			recipes, exists := recipeMap[current.Element]
			if exists {
				for _, recipe := range recipes {
					for _, ingredient := range recipe {
						ingredientElement, exists := elementMap[ingredient]
						if !exists {
							continue
						}
						
						ingredientTier := ingredientElement.Tier
						
						if !backwardVisited[ingredient] || ingredientTier < bestTierPath[ingredient] {
							// Only add if not visited or if we found a better tier path
							otherIngredient := recipe[0]
							if recipe[0] == ingredient {
								otherIngredient = recipe[1]
							}
							
							backwardQueue.PushBack(&models.Node{
								Element: ingredient,
								Path:    append(current.Path, current.Element, otherIngredient),
								Tier:    ingredientTier,
							})
							backwardVisited[ingredient] = true
							backwardPath[ingredient] = []string{current.Element, otherIngredient}
							bestTierPath[ingredient] = ingredientTier
							
							// Check if we've reached an element discovered by forward search
							if forwardVisited[ingredient] {
								meetingPoints = append(meetingPoints, MeetingPoint{
									Element: ingredient,
									Forward: false,
									Path:    append(current.Path, current.Element, otherIngredient),
									Tier:    ingredientTier,
								})
							}
							
							// Check if this ingredient is a basic element
							for _, basic := range basicElements {
								if ingredient == basic {
									// Found a direct path from target to basic element
									meetingPoints = append(meetingPoints, MeetingPoint{
										Element: basic,
										Forward: false,
										Path:    append(current.Path, current.Element, otherIngredient),
										Tier:    0, // Basic elements are tier 0
									})
								}
							}
						}
					}
				}
			}
		}
		
		// If we found meeting points, we can stop the search
		if len(meetingPoints) > 0 {
			break
		}
	}

	// If no meeting points found, no path was found
	if len(meetingPoints) == 0 {
		return models.RecipeTree{}, time.Since(startTime).Seconds(), nodesVisited
	}

	// Choose the meeting point with the lowest tier sum for the path
	var bestMeetingPoint MeetingPoint
	bestTierSum := 1000000 // Initialize with a large number
	
	for _, mp := range meetingPoints {
		// For basic elements, always choose them
		if mp.Tier == 0 {
			bestMeetingPoint = mp
			break
		}
		
		if mp.Tier < bestTierSum {
			bestTierSum = mp.Tier
			bestMeetingPoint = mp
		}
	}

	// We found a path, now build the recipe tree
	recipeTree := buildRecipeTree(targetElement, bestMeetingPoint.Element, forwardPath, backwardPath, recipeMap, elementMap)
	return recipeTree, time.Since(startTime).Seconds(), nodesVisited
}

// buildRecipeTree builds a recipe tree from the meeting point
func buildRecipeTree(targetElement, meetingPoint string, forwardPath, backwardPath map[string][]string, recipeMap map[string][][]string, elementMap map[string]models.Element) models.RecipeTree {
	// Use a visited map to prevent cycles
	visited := make(map[string]bool)
	
	// Find the best recipe for the target element
	lowestTierSum := 1000000
	
	recipes, exists := recipeMap[targetElement]
	if exists {
		for _, recipe := range recipes {
			tierSum := 0
			for _, ingredient := range recipe {
				if el, ok := elementMap[ingredient]; ok {
					tierSum += el.Tier
				}
			}
			
			if tierSum < lowestTierSum {
				lowestTierSum = tierSum
			}
		}
	}
	
	// Recursively build the tree using the best recipe
	return buildRecipeTreeHelper(targetElement, visited, forwardPath, backwardPath, recipeMap, elementMap)
}

func buildRecipeTreeHelper(element string, visited map[string]bool, forwardPath, backwardPath map[string][]string, recipeMap map[string][][]string, elementMap map[string]models.Element) models.RecipeTree {
	// Prevent infinite recursion
	if visited[element] {
		return models.RecipeTree{
			Root:     element,
			Left:     "",
			Right:    "",
			Tier:     fmt.Sprintf("%d", elementMap[element].Tier),
			Children: []models.RecipeTree{},
		}
	}
	
	visited[element] = true
	
	// Get the tier for the element
	var tier int
	if el, exists := elementMap[element]; exists {
		tier = el.Tier
	}
	
	// Create recipe tree
	recipeTree := models.RecipeTree{
		Root:     element,
		Left:     "",
		Right:    "",
		Tier:     fmt.Sprintf("%d", tier),
		Children: []models.RecipeTree{},
	}
	
	// If this is a basic element (tier 0), return as is
	if tier == 0 {
		return recipeTree
	}
	
	// Find the best recipe with lowest tier ingredients
	var bestRecipe []string
	lowestTierSum := 1000000
	
	recipes, exists := recipeMap[element]
	if exists {
		for _, recipe := range recipes {
			tierSum := 0
			for _, ingredient := range recipe {
				if el, ok := elementMap[ingredient]; ok {
					tierSum += el.Tier
				}
			}
			
			if tierSum < lowestTierSum {
				lowestTierSum = tierSum
				bestRecipe = recipe
			}
		}
	}
	
	// If no recipe found, return as is
	if len(bestRecipe) == 0 {
		return recipeTree
	}
	
	// Set the ingredients in the recipe tree
	recipeTree.Left = bestRecipe[0]
	recipeTree.Right = bestRecipe[1]
	
	// Now recursively build children for both ingredients
	leftChild := buildRecipeTreeHelper(bestRecipe[0], visited, forwardPath, backwardPath, recipeMap, elementMap)
	rightChild := buildRecipeTreeHelper(bestRecipe[1], visited, forwardPath, backwardPath, recipeMap, elementMap)
	
	// Add children to the recipe tree
	recipeTree.Children = append(recipeTree.Children, leftChild, rightChild)
	
	return recipeTree
}

// createRecipeMap builds the recipe map from the element list
func createRecipeMap(elements []models.Element) map[string][][]string {
	recipeMap := make(map[string][][]string)
	
	for _, element := range elements {
		// Handle elements with recipes
		if len(element.Recipes) == 2 {
			// Add to recipe map
			if _, exists := recipeMap[element.Name]; !exists {
				recipeMap[element.Name] = make([][]string, 0)
			}
			recipeMap[element.Name] = append(recipeMap[element.Name], element.Recipes)
		}
	}
	
	return recipeMap
}