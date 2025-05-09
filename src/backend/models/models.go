package models

// Struktur data elemen sesuai dengan output dari scraper
type Element struct {
    Name    string   `json:"name"`
    Recipes []string `json:"recipes"`
    Tier    int      `json:"tier"`
}

// Struktur untuk visualisasi recipe tree
type RecipeNode struct {
    Element    string       `json:"element"`
    Components []RecipeNode `json:"components,omitempty"`
    IsBasic    bool         `json:"isBasic"`
}

// Struktur untuk respons API
type SearchResponse struct {
    Recipes      []RecipeNode `json:"recipes"`
    TimeElapsed  float64      `json:"timeElapsed"`
    NodesVisited int          `json:"nodesVisited"`
}

// Struktur untuk hasil pencarian recipe
type SearchResult struct {
    Recipes      []RecipeNode `json:"recipes"`
    NodesVisited int          `json:"nodesVisited"`
    TimeElapsed  float64      `json:"timeElapsed,omitempty"`
    Error        string       `json:"error,omitempty"`
}

// Struktur untuk request API
type SearchRequest struct {
    TargetElement  string `json:"targetElement"`
    Algorithm      string `json:"algorithm"` // "bfs", "dfs", atau "bidirectional"
    MultipleRecipes bool   `json:"multipleRecipes"`
    MaxRecipes     int    `json:"maxRecipes"`
}

// Helper function untuk memeriksa elemen dasar
func IsBasicElement(element Element) bool {
    return element.Tier == 0
}

// Helper function untuk memeriksa elemen dasar berdasarkan nama
func IsBasicElementByName(name string, elements []Element) bool {
    for _, el := range elements {
        if el.Name == name && el.Tier == 0 {
            return true
        }
    }
    return false
}

// Helper function untuk mendapatkan elemen dari nama
func GetElementByName(name string, elements []Element) (Element, bool) {
    for _, el := range elements {
        if el.Name == name {
            return el, true
        }
    }
    return Element{}, false
}

// Helper function untuk membuat map elemen untuk akses cepat
func CreateElementMap(elements []Element) map[string]Element {
    elementMap := make(map[string]Element)
    for _, el := range elements {
        elementMap[el.Name] = el
    }
    return elementMap
}