package main

import (
    "fmt"
    "os"
    "encoding/json"
    
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    
    "Tubes2_alchendol/api"
    "Tubes2_alchendol/search"
)

func main() {
    // Setup API Server
    r := gin.Default()
    
    // Setup CORS with more permissive settings
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
        ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
        AllowCredentials: true,
        MaxAge:           86400, // 24 hours
    }))
    
    // Add a health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "message": "Server is running",
        })
    })
    
    // Setup API routes
    r.GET("/api/search", api.SearchHandler)
    
    // Terminal-based interface
    if len(os.Args) > 1 && os.Args[1] == "--cli" {
        runCLI()
    } else {
        // Run server on a specific host and port
        port := os.Getenv("PORT")
        if port == "" {
            port = "8080"
        }
        fmt.Printf("Server starting on port %s...\n", port)
        r.Run(":" + port)
    }
}

func runCLI() {
    // Load element data from JSON
    elements, err := api.LoadRecipeData()
    if err != nil {
        fmt.Printf("Error membaca elemen: %v\n", err)
        return
    }

    // Pilih algoritma dan target dari input terminal
    var algo, target, multiple string
    fmt.Print("Pilih algoritma (DFS / BIDIRECTIONAL): ")
    fmt.Scanln(&algo)
    fmt.Print("Masukkan mode multiple recipe (y/n): ")
    fmt.Scanln(&multiple)
    fmt.Print("Masukkan nama elemen target: ")
    fmt.Scanln(&target)

    switch multiple {
    case "y":
        var maxRecipes int
        fmt.Print("Masukkan jumlah maksimal resep yang ingin dicari (max 100): ")
        fmt.Scanln(&maxRecipes)
        recipes, timeTaken, nodes := search.MultipleDFS(target, elements, maxRecipes)
        fmt.Printf("Multiple DFS selesai dalam %.6f detik, %d node dikunjungi\n", timeTaken, nodes)
        
        // Output all recipes as a single JSON array
        output := map[string]interface{}{
            "roots": recipes,
            "timeElapsed": timeTaken,
            "nodesVisited": nodes,
        }
        outputResult(output)
        return
    case "n":
        // Do nothing, continue to single recipe search
    default:
        fmt.Println("Pilihan tidak valid. Gunakan 'y' untuk multiple atau 'n' untuk single.")
        return
    }
    // Hasil pencarian
    switch algo {
    case "DFS":
        recipes, timeTaken, nodes := search.DFS(target, elements)
        fmt.Printf("DFS selesai dalam %.6f detik, %d node dikunjungi\n", timeTaken, nodes)
        outputResult(recipes)
    case "BIDIRECTIONAL":
        recipes, timeTaken, nodes := search.BidirectionalSearch(target, elements)
        fmt.Printf("Bidirectional selesai dalam %.6f detik, %d node dikunjungi\n", timeTaken, nodes)
        outputResult(recipes)
    default:
        fmt.Println("Algoritma tidak dikenali. Gunakan 'DFS' atau 'BIDIRECTIONAL'.")
    }
}

// Fungsi bantu untuk mencetak hasil dalam format JSON
func outputResult(data interface{}) {
    jsonResult, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        fmt.Println("Error saat konversi ke JSON:", err)
        os.Exit(1)
    }
    fmt.Println(string(jsonResult))
}