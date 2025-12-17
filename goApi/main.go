package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type RepoData struct {
	Name      string         `json:"name"`
	Readme    string         `json:"readme"`
	Languages map[string]int `json:"languages"`
}

func fetchRepoData(repoName string) RepoData {
	repo := RepoData{
		Name:      repoName,
		Languages: make(map[string]int),
	}

	// GET Readme
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

	// GET ALL languages
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

func main() {
	godotenv.Load()
	r := gin.Default()

	// Get allowed origins from env
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	// Add CORS middleware with env-based origins
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	r.GET("/api/repos", func(c *gin.Context) {
		allRepos, err := getAllRepos()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch repos"})
			return
		}
		c.JSON(200, allRepos)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
