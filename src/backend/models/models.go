package models

type Element struct {
    Name    string   `json:"name"`
    Recipes []string `json:"recipes"`
    Tier    int      `json:"tier"`
}

type RecipeNode struct {
    Element    string       `json:"element"`
    Components []RecipeNode `json:"components,omitempty"`
    IsBasic    bool         `json:"isBasic"`
}

type RecipeTree struct {
    Root     string       `json:"root"`
    Left     string       `json:"Left"`
    Right    string       `json:"Right"`
    Tier     string       `json:"Tier"`
    Children []RecipeTree `json:"children"`
}

type SearchResponse struct {
    Recipes      []RecipeNode `json:"recipes"`
    TimeElapsed  float64      `json:"timeElapsed"`
    NodesVisited int          `json:"nodesVisited"`
}

type SearchResult struct {
    RecipeTree   RecipeTree   `json:"recipeTree,omitempty"` 
}

type SearchRequest struct {
    TargetElement  string `json:"targetElement"`
    Algorithm      string `json:"algorithm"` 
    MultipleRecipes bool   `json:"multipleRecipes"`
    MaxRecipes     int    `json:"maxRecipes"`
}

type Node struct {
	Element string
	Path    []string
	Parent  *Node
	Tier    int
}

func IsBasicElement(element Element) bool {
    return element.Tier == 0
}

func IsBasicElementByName(name string, elements []Element) bool {
    for _, el := range elements {
        if el.Name == name && el.Tier == 0 {
            return true
        }
    }
    return false
}

func GetElementByName(name string, elements []Element) (Element, bool) {
    for _, el := range elements {
        if el.Name == name {
            return el, true
        }
    }
    return Element{}, false
}

func CreateElementMap(elements []Element) map[string]Element {
    elementMap := make(map[string]Element)
    for _, el := range elements {
        elementMap[el.Name] = el
    }
    return elementMap
}