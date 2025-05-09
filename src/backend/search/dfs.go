package search

import (
	"fmt"
	"sync"
	"time"

	"Tubes2_alchendol/models"
)

// Cache structure to store already computed recipe trees
type RecipeCache struct {
	cache map[string]models.RecipeNode
	mutex sync.RWMutex
}

// Create a new recipe cache
func NewRecipeCache() *RecipeCache {
	return &RecipeCache{
		cache: make(map[string]models.RecipeNode),
	}
}

// Get a recipe from cache
func (c *RecipeCache) Get(elementName string) (models.RecipeNode, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	node, exists := c.cache[elementName]
	return node, exists
}

// Store a recipe in cache
func (c *RecipeCache) Set(elementName string, node models.RecipeNode) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[elementName] = node
}

// Fungsi untuk membuat map elemen yang difilter sesuai dengan aturan tier
func CreateFilteredElementMap(elements []models.Element, target string) (map[string][]models.Element, bool) {
	targetFound := false
	elementMap := make(map[string][]models.Element)
	targetTier := -1
	
	// Pertama, temukan target dan tier-nya
	for _, el := range elements {
		if el.Name == target {
			targetFound = true
			// Update targetTier dengan tier tertinggi dari elemen target
			if el.Tier > targetTier {
				targetTier = el.Tier
			}
			// Tambahkan target ke map
			elementMap[el.Name] = append(elementMap[el.Name], el)
		}
	}
	
	// Jika target tidak ditemukan, return kosong
	if !targetFound {
		fmt.Printf("DEBUG: Target '%s' not found in elements\n", target)
		return elementMap, false
	}
	
	// Kemudian, tambahkan elemen non-dasar dengan tier lebih rendah dari target
	for _, el := range elements {
		// Skip elemen dengan nama target (sudah ditambahkan) dan elemen dasar (tier 0)
		if el.Name == target || el.Tier == 0 {
			continue
		}
		
		// Tambahkan elemen dengan tier lebih rendah dari target
		if el.Tier < targetTier {
			elementMap[el.Name] = append(elementMap[el.Name], el)
		}
	}
	
	return elementMap, true
}

// Add basic elements to the element map
func AddBasicElements(elementMap map[string][]models.Element) {
	basicElements := []string{"Air", "Earth", "Fire", "Water"}
	for _, name := range basicElements {
		elementMap[name] = []models.Element{{
			Name:    name,
			Recipes: []string{},
			Tier:    0,
		}}
	}
}

// Struktur untuk menyimpan state pencarian DFS
type DFSState struct {
	Target       string
	ElementMap   map[string][]models.Element
	Path         []string  // Path yang diikuti dalam pencarian ini
	NodesVisited int
	Cache        *RecipeCache // Cache untuk hasil pencarian
}

// Check if element is a basic element
func IsBasicElement(elementName string) bool {
	return elementName == "Air" || elementName == "Earth" || elementName == "Fire" || elementName == "Water"
}

// Fungsi utama untuk DFS - mencari satu recipe terpendek
func DFS(target string, elements []models.Element) models.SearchResult {
	// Inisialisasi waktu mulai
	startTime := time.Now()
	
	// Filter elemen hanya sampai target
	elementMap, targetFound := CreateFilteredElementMap(elements, target)
	if !targetFound {
		return models.SearchResult{
			NodesVisited: 0,
			Error:        fmt.Sprintf("Elemen '%s' tidak ditemukan", target),
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}
	
	// Add basic elements to map
	AddBasicElements(elementMap)
	
	// Periksa apakah target adalah elemen dasar
	if IsBasicElement(target) {
		node := models.RecipeNode{
			Element: target,
			IsBasic: true,
		}
		return models.SearchResult{
			Recipes:      []models.RecipeNode{node},
			NodesVisited: 1,
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}
	
	// Inisialisasi state dengan cache
	state := &DFSState{
		Target:       target,
		ElementMap:   elementMap,
		Path:         make([]string, 0),
		NodesVisited: 0,
		Cache:        NewRecipeCache(),
	}
	
	// Mulai pencarian DFS dengan cache
	recipeNode, found := dfsSearchRecipeWithCache(state, target)
	if !found {
		return models.SearchResult{
			NodesVisited: state.NodesVisited,
			Error:        fmt.Sprintf("Tidak dapat menemukan resep untuk '%s'", target),
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}

	return models.SearchResult{
		Recipes:      []models.RecipeNode{recipeNode},
		NodesVisited: state.NodesVisited,
		TimeElapsed:  time.Since(startTime).Seconds(),
	}
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

// Fungsi rekursif untuk mencari resep menggunakan DFS dengan cache
func dfsSearchRecipeWithCache(state *DFSState, elementName string) (models.RecipeNode, bool) {
	// Mencatat kunjungan ke node
	state.NodesVisited++
	
	// Check if the element is in cache
	if cachedNode, exists := state.Cache.Get(elementName); exists {
		return cachedNode, true
	}
	
	// If this is a basic element, return directly
	if IsBasicElement(elementName) {
		node := models.RecipeNode{
			Element: elementName,
			IsBasic: true,
		}
		// Cache the result
		state.Cache.Set(elementName, node)
		return node, true
	}
	
	// Tambahkan elemen ke path untuk deteksi siklus
	if isInPath(state.Path, elementName) {
		return models.RecipeNode{}, false
	}
	state.Path = append(state.Path, elementName)
	defer func() {
		// Hapus elemen dari path sebelum return
		if len(state.Path) > 0 {
			state.Path = state.Path[:len(state.Path)-1]
		}
	}()
	
	// Dapatkan elemen dari map
	elementSlice, found := state.ElementMap[elementName]
	if !found || len(elementSlice) == 0 {
		return models.RecipeNode{}, false
	}
	
	// Coba setiap resep dari elemen
	for _, element := range elementSlice {
		// Pastikan array recipes memiliki tepat 2 komponen
		if len(element.Recipes) != 2 {
			continue
		}
		
		comp1 := element.Recipes[0]
		comp2 := element.Recipes[1]
		
		// Cari resep untuk komponen pertama
		comp1Node, found1 := dfsSearchRecipeWithCache(state, comp1)
		if !found1 {
			continue
		}
		
		// Cari resep untuk komponen kedua
		comp2Node, found2 := dfsSearchRecipeWithCache(state, comp2)
		if !found2 {
			continue
		}
		
		// Buat node resep dengan kedua komponen
		node := models.RecipeNode{
			Element:    elementName,
			Components: []models.RecipeNode{comp1Node, comp2Node},
			IsBasic:    false,
		}
		
		// Cache the successful result
		state.Cache.Set(elementName, node)
		
		return node, true
	}
	
	// Jika tidak ada resep yang berhasil ditemukan
	return models.RecipeNode{}, false
}

// Fungsi untuk multiple recipe dengan DFS menggunakan goroutines
func DFSMultiple(target string, elements []models.Element, maxRecipes int) models.SearchResult {
	startTime := time.Now()
	
	// Filter elemen hanya sampai target
	elementMap, targetFound := CreateFilteredElementMap(elements, target)
	if !targetFound {
		return models.SearchResult{
			NodesVisited: 0,
			Error:        fmt.Sprintf("Elemen '%s' tidak ditemukan", target),
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}
	
	// Add basic elements to map
	AddBasicElements(elementMap)
	
	// Periksa apakah target adalah elemen dasar
	if IsBasicElement(target) {
		node := models.RecipeNode{
			Element: target,
			IsBasic: true,
		}
		return models.SearchResult{
			Recipes:      []models.RecipeNode{node},
			NodesVisited: 1,
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}
	
	// Inisialisasi state dengan cache bersama
	sharedCache := NewRecipeCache()
	nodesVisited := 0
	
	// Channel untuk mengumpulkan hasil
	resultChan := make(chan models.RecipeNode, maxRecipes)
	
	// Wait group untuk menunggu semua goroutine selesai
	var wg sync.WaitGroup
	
	// Mutex untuk mengakses counter nodesVisited
	var visitMutex sync.Mutex
	
	// Slice untuk menyimpan hasil akhir
	var recipes []models.RecipeNode
	
	// Fungsi untuk memproses satu elemen dan recipe-nya
	processElement := func(element models.Element) {
		defer wg.Done()
		
		// Buat state baru untuk setiap goroutine dengan cache bersama
		state := &DFSState{
			Target:       target,
			ElementMap:   elementMap,
			Path:         make([]string, 0),
			NodesVisited: 0,
			Cache:        sharedCache,
		}
		
		// Cari recipe untuk elemen ini
		recipeNode, found := exploreRecipe(state, element)
		
		// Update counter nodesVisited secara thread-safe
		visitMutex.Lock()
		nodesVisited += state.NodesVisited
		visitMutex.Unlock()
		
		// Jika ditemukan recipe, kirim ke channel
		if found {
			select {
			case resultChan <- recipeNode:
				// Recipe berhasil dikirim
			default:
				// Channel penuh, abaikan
			}
		}
	}
	
	// Dapatkan semua elemen target dari elementMap
	targetElements, ok := elementMap[target]
	if !ok || len(targetElements) == 0 {
		return models.SearchResult{
			NodesVisited: 0,
			Error:        fmt.Sprintf("Tidak dapat menemukan resep untuk '%s'", target),
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}
	
	// Jalankan goroutine untuk setiap elemen target
	for _, element := range targetElements {
		wg.Add(1)
		go processElement(element)
	}
	
	// Goroutine untuk menutup channel setelah semua goroutine selesai
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Kumpulkan hasil dari channel
	for recipe := range resultChan {
		recipes = append(recipes, recipe)
		if len(recipes) >= maxRecipes {
			break
		}
	}
	
	// Jika tidak ada recipe yang ditemukan
	if len(recipes) == 0 {
		return models.SearchResult{
			NodesVisited: nodesVisited,
			Error:        fmt.Sprintf("Tidak dapat menemukan resep untuk '%s'", target),
			TimeElapsed:  time.Since(startTime).Seconds(),
		}
	}
	
	return models.SearchResult{
		Recipes:      recipes,
		NodesVisited: nodesVisited,
		TimeElapsed:  time.Since(startTime).Seconds(),
	}
}

// Fungsi untuk menjelajahi satu recipe secara terpisah
func exploreRecipe(state *DFSState, element models.Element) (models.RecipeNode, bool) {
	if len(element.Recipes) != 2 {
		return models.RecipeNode{}, false
	}
	
	comp1 := element.Recipes[0]
	comp2 := element.Recipes[1]
	
	// Cari resep untuk komponen pertama
	comp1Node, found1 := dfsSearchRecipeWithCache(state, comp1)
	if !found1 {
		return models.RecipeNode{}, false
	}
	
	// Cari resep untuk komponen kedua
	comp2Node, found2 := dfsSearchRecipeWithCache(state, comp2)
	if !found2 {
		return models.RecipeNode{}, false
	}
	
	// Buat node resep dengan kedua komponen
	return models.RecipeNode{
		Element:    element.Name,
		Components: []models.RecipeNode{comp1Node, comp2Node},
		IsBasic:    false,
	}, true
}