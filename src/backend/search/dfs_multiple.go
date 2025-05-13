package search

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"runtime"
	"context"
	"Tubes2_alchendol/models"
)

// MultipleDFS searches for multiple recipes using concurrent DFS with multithreading
func MultipleDFS(target string, elements []models.Element, maxRecipes int) ([]models.RecipeTree, float64, int) {
    startTime := time.Now()
    
    // Tingkatkan timeout untuk elemen kompleks
    timeoutDuration := 120 * time.Second // Tingkatkan dari 15 atau 60 detik
    for _, el := range elements {
        if el.Name == target {
            if el.Tier > 5 {
                timeoutDuration = 180 * time.Second // 3 menit untuk elemen tier tinggi
            } else if el.Tier > 3 {
                timeoutDuration = 120 * time.Second // 2 menit untuk tier menengah
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
    
    // Gunakan atomic counter untuk tracking node yang dikunjungi
    var totalNodesVisited int64 = 0
    
    // Hasil recipes dan map untuk menjaga keunikan
    var results []models.RecipeTree
    var uniqueRecipes = make(map[string]bool)
    var resultsMutex sync.Mutex
    
    // Dapatkan semua kombinasi langsung untuk target
    var allCombinations [][]string
    targetElements, found := elementMap[target]
    if !found || len(targetElements) == 0 {
        return []models.RecipeTree{}, time.Since(startTime).Seconds(), 0
    }
    
    // Kumpulkan semua kombinasi untuk target
    for _, element := range targetElements {
        if len(element.Recipes) == 2 {
            allCombinations = append(allCombinations, element.Recipes)
        }
    }
    
    fmt.Printf("Found %d direct combinations for %s\n", len(allCombinations), target)
    
    // Channel untuk mengumpulkan hasil
    resultChan := make(chan models.RecipeTree, 1000) // Buffer besar untuk menghindari blocking
    
    // Channel untuk sinyal done
    doneChan := make(chan struct{})
    
    // Mulai collector goroutine - mirip dengan kode kelompok lain
    go func() {
        treesProcessed := 0
        for tree := range resultChan {
            if tree.Root == "" {
                continue
            }
            
            // Buat key unik untuk tree
            treeKey := generateCompleteTreeKey(tree)
            
            resultsMutex.Lock()
            
            // Cek jika sudah mencapai max
            if len(results) >= maxRecipes {
                resultsMutex.Unlock()
                continue
            }
            
            // Hanya tambahkan jika belum ada
            if !uniqueRecipes[treeKey] {
                uniqueRecipes[treeKey] = true
                results = append(results, tree)
                
                // Log progress
                treesProcessed++
                if treesProcessed % 10 == 0 || treesProcessed < 10 {
                    fmt.Printf("Added recipe %d/%d: %s = %s + %s (NodesVisited: %d)\n", 
                        len(results), maxRecipes, tree.Root, tree.Left, tree.Right, 
                        atomic.LoadInt64(&totalNodesVisited))
                }
                
                // Signal done jika sudah mencapai max
                if len(results) >= maxRecipes {
                    select {
                    case doneChan <- struct{}{}:
                    default:
                    }
                }
            }
            
            resultsMutex.Unlock()
        }
    }()
    
    // Tangani semua kombinasi menggunakan worker pool
    combinationChan := make(chan []string, len(allCombinations))
    var wg sync.WaitGroup
    
    // Jumlah worker berdasarkan CPU
    workerCount := runtime.NumCPU()
    fmt.Printf("Starting %d workers for processing combinations\n", workerCount)
    
    // Start worker goroutines
    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            
            // Buat cache lokal untuk worker ini
            localCache := NewRecipeCache()
            
            for combo := range combinationChan {
                // Check for context done or reached max recipes
                select {
                case <-ctx.Done():
                    return
                case <-doneChan:
                    return
                default:
                    // Continue processing
                }
                
                // Get component elements
                comp1 := combo[0]
                comp2 := combo[1]
                
                // Skip jika tier violation
                targetTier := getElementTier(target, elementMap)
                comp1Tier := getElementTier(comp1, elementMap)
                comp2Tier := getElementTier(comp2, elementMap)
                
                if comp1Tier >= targetTier || comp2Tier >= targetTier {
                    continue
                }
                
                // Dapatkan semua variasi tree untuk comp1
                variations1 := processComponentVariations(comp1, elementMap, localCache, &totalNodesVisited)
                
                // Dapatkan semua variasi tree untuk comp2
                variations2 := processComponentVariations(comp2, elementMap, localCache, &totalNodesVisited)
                
                // Kombinasikan semua variasi untuk membuat berbagai recipe tree
                for _, var1 := range variations1 {
                    // Check for done signal periodically
                    select {
                    case <-ctx.Done():
                        return
                    case <-doneChan:
                        return
                    default:
                        // Continue
                    }
                    
                    for _, var2 := range variations2 {
                        // Buat recipe tree dengan variasi ini
                        recipeTree := models.RecipeTree{
                            Root:     target,
                            Left:     comp1,
                            Right:    comp2,
                            Tier:     fmt.Sprintf("%d", targetTier),
                            Children: []models.RecipeTree{var1, var2},
                        }
                        
                        // Kirim ke channel
                        select {
                        case resultChan <- recipeTree:
                            // Increment counter for each tree added
                            atomic.AddInt64(&totalNodesVisited, 1)
                        case <-ctx.Done():
                            return
                        case <-doneChan:
                            return
                        }
                    }
                }
            }
        }(i)
    }
    
    // Kirim kombinasi ke channel untuk diproses
    for _, combo := range allCombinations {
        combinationChan <- combo
    }
    close(combinationChan)
    
    // Tunggu semua worker selesai
    go func() {
        wg.Wait()
        close(resultChan)
    }()
    
    // Tunggu sampai context done, atau doneChan signal, atau semua worker selesai
    select {
    case <-ctx.Done():
        fmt.Println("Context timeout reached, using results gathered so far")
    case <-doneChan:
        fmt.Printf("Found all %d requested recipes\n", maxRecipes)
    case <-func() chan struct{} {
        ch := make(chan struct{})
        go func() {
            wg.Wait()
            close(ch)
        }()
        return ch
    }():
        fmt.Println("All combinations processed")
    }
    
    // Tunggu sedikit untuk collector goroutine untuk menyelesaikan
    cleanup := time.After(2 * time.Second)
    <-cleanup
    
    finalNodeCount := atomic.LoadInt64(&totalNodesVisited)
    
    fmt.Printf("MultipleDFS returning %d recipes for %s with %d nodes visited\n", 
        len(results), target, finalNodeCount)
    
    return results, time.Since(startTime).Seconds(), int(finalNodeCount)
}

// Fungsi helper untuk memproses semua variasi komponen
func processComponentVariations(component string, elementMap map[string][]models.Element, 
    cache *RecipeCache, totalNodesVisited *int64) []models.RecipeTree {
    
    // Increment visit counter
    atomic.AddInt64(totalNodesVisited, 1)
    
    // If basic element, return direct leaf
    if IsBasicElement(component) {
        return []models.RecipeTree{
            {
                Root:     component,
                Left:     "",
                Right:    "",
                Tier:     "0",
                Children: []models.RecipeTree{},
            },
        }
    }
    
    // Check cache
    if cachedNode, exists := cache.Get(component); exists {
        return []models.RecipeTree{ConvertToRecipeTree(cachedNode, elementMap)}
    }
    
    // Get component tier
    compTier := getElementTier(component, elementMap)
    
    // Get all recipes
    componentElements, found := elementMap[component]
    if !found || len(componentElements) == 0 {
        return []models.RecipeTree{}
    }
    
    // Collect all variations
    var variations []models.RecipeTree
    processedCombos := make(map[string]bool)
    
    // Limit max variations per element to avoid exponential explosion
    maxVariationsPerElement := 50 // Tingkatkan dari 5 ke 50
    
    for _, element := range componentElements {
        if len(variations) >= maxVariationsPerElement {
            break
        }
        
        if len(element.Recipes) != 2 {
            continue
        }
        
        subComp1 := element.Recipes[0]
        subComp2 := element.Recipes[1]
        
        // Check tier constraints
        subComp1Tier := getElementTier(subComp1, elementMap)
        subComp2Tier := getElementTier(subComp2, elementMap)
        
        if subComp1Tier >= compTier || subComp2Tier >= compTier {
            continue
        }
        
        // Create unique key for this combo
        var comboKey string
        if subComp1 < subComp2 {
            comboKey = subComp1 + "+" + subComp2
        } else {
            comboKey = subComp2 + "+" + subComp1
        }
        
        // Skip if already processed
        if processedCombos[comboKey] {
            continue
        }
        processedCombos[comboKey] = true
        
        // Recursively get variations for subcomponents
        subComp1Trees := processComponentVariations(subComp1, elementMap, cache, totalNodesVisited)
        if len(subComp1Trees) == 0 {
            continue
        }
        
        subComp2Trees := processComponentVariations(subComp2, elementMap, cache, totalNodesVisited)
        if len(subComp2Trees) == 0 {
            continue
        }
        
        // Combine variations
        for i, comp1Tree := range subComp1Trees {
            if len(variations) >= maxVariationsPerElement {
                break
            }
            
            for j, comp2Tree := range subComp2Trees {
                if len(variations) >= maxVariationsPerElement {
                    break
                }
                
                // Only take a subset of combinations to avoid explosion
                if i*len(subComp2Trees) + j >= maxVariationsPerElement {
                    break
                }
                
                tree := models.RecipeTree{
                    Root:     component,
                    Left:     subComp1,
                    Right:    subComp2,
                    Tier:     fmt.Sprintf("%d", compTier),
                    Children: []models.RecipeTree{comp1Tree, comp2Tree},
                }
                
                variations = append(variations, tree)
            }
        }
    }
    
    // Cache the result (first variation only)
    if len(variations) > 0 {
        node := convertRecipeTreeToNode(variations[0])
        cache.Set(component, node)
    }
    
    return variations
}

// Helper untuk mengkonversi RecipeTree ke RecipeNode untuk caching
func convertRecipeTreeToNode(tree models.RecipeTree) models.RecipeNode {
    if len(tree.Children) == 0 {
        return models.RecipeNode{
            Element: tree.Root,
            IsBasic: true,
        }
    }
    
    components := make([]models.RecipeNode, 0, 2)
    if len(tree.Children) >= 2 {
        components = append(components, 
            convertRecipeTreeToNode(tree.Children[0]),
            convertRecipeTreeToNode(tree.Children[1]))
    }
    
    return models.RecipeNode{
        Element:    tree.Root,
        Components: components,
        IsBasic:    false,
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