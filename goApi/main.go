package main
import (
    "bytes"
    "encoding/base64"
    "encoding/json"
    "os"
    "net/http"
    "sync"
    "strings"
    "github.com/joho/godotenv"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
)

type RepoData struct {
    Name      string            `json:"name"`
    Readme    string            `json:"readme"`
    Languages map[string]int    `json:"languages"`
}

var (
    cachedRepos []RepoData
    cacheMutex  sync.RWMutex
    isCached    bool
)

func fetchRepoData(repoName string) RepoData {
    repo := RepoData{
        Name:      repoName,
        Languages: make(map[string]int),
    }
    
    respReadme, err := http.Get("https://api.github.com/repos/BraedynL0530/" + repoName + "/readme")
    if err == nil {
        defer respReadme.Body.Close()
        var readmeData map[string]interface{}
        if json.NewDecoder(respReadme.Body).Decode(&readmeData) == nil {
            if content, ok := readmeData["content"].(string); ok {
                decoded, _ := decodeBase64(content)
                repo.Readme = decoded
            }
        }
    }
    
    respLanguage, err := http.Get("https://api.github.com/repos/BraedynL0530/" + repoName + "/languages")
    if err == nil {
        defer respLanguage.Body.Close()
        var languageData map[string]interface{}
        if json.NewDecoder(respLanguage.Body).Decode(&languageData) == nil {
            for lang, bytes := range languageData {
                if b, ok := bytes.(float64); ok {
                    repo.Languages[lang] = int(b)
                }
            }
        }
    }
    
    return repo
}

func decodeBase64(encoded string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

func getAllRepos() ([]RepoData, error) {
    resp, err := http.Get("https://api.github.com/users/BraedynL0530/repos")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var repos []map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&repos)

    var wg sync.WaitGroup
    results := make(chan RepoData, len(repos))

    for _, r := range repos {
        name := r["name"].(string)
        wg.Add(1)
        go func(repoName string) {
            defer wg.Done()
            results <- fetchRepoData(repoName)
        }(name)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    allRepos := []RepoData{}
    for repoData := range results {
        allRepos = append(allRepos, repoData)
    }

    return allRepos, nil
}

func sendToDjango(repos []RepoData) error {
    djangoURL := os.Getenv("DJANGO_API_URL")
    
    jsonData, err := json.Marshal(repos)
    if err != nil {
        return err
    }

    resp, err := http.Post(
        djangoURL,
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func main() {
    godotenv.Load()
    r := gin.Default()
    
    allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
    
    r.Use(cors.New(cors.Config{
        AllowOrigins:     allowedOrigins,
        AllowMethods:     []string{"GET", "POST"},
        AllowHeaders:     []string{"Origin", "Content-Type"},
        AllowCredentials: true,
    }))
    
    // Endpoint to get raw repos (for testing)
    r.GET("/api/repos", func(c *gin.Context) {
        cacheMutex.RLock()
        if isCached && len(cachedRepos) > 0 {
            repos := cachedRepos
            cacheMutex.RUnlock()
            c.JSON(200, gin.H{"data": repos, "cached": true})
            return
        }
        cacheMutex.RUnlock()

        allRepos, err := getAllRepos()
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to fetch repos"})
            return
        }

        cacheMutex.Lock()
        cachedRepos = allRepos
        isCached = true
        cacheMutex.Unlock()

        c.JSON(200, gin.H{"data": allRepos, "cached": false})
    })

    // Sync to Django for NLP processing
    r.POST("/api/sync-to-django", func(c *gin.Context) {
        allRepos, err := getAllRepos()
        if err != nil {
            c.JSON(500, gin.H{"error": "failed to fetch repos"})
            return
        }

        if err := sendToDjango(allRepos); err != nil {
            c.JSON(500, gin.H{"error": "failed to send to django", "details": err.Error()})
            return
        }

        cacheMutex.Lock()
        cachedRepos = allRepos
        isCached = true
        cacheMutex.Unlock()

        c.JSON(200, gin.H{"message": "synced to django", "count": len(allRepos)})
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}