package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type RepoData struct {
	Name      string         `json:"name"`
	Readme    string         `json:"readme"`
	Languages map[string]int `json:"languages"`
}

var (
	redisClient *redis.Client
	ctx         = context.Background()
	cacheTTL    = 1 * time.Hour
)

func init() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test connection
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}

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

// Cache operations
func getCachedRepos() ([]RepoData, error) {
	val, err := redisClient.Get(ctx, "repos:all").Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err // Redis error
	}

	var repos []RepoData
	if err := json.Unmarshal([]byte(val), &repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func setCachedRepos(repos []RepoData) error {
	jsonData, err := json.Marshal(repos)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, "repos:all", string(jsonData), cacheTTL).Err()
}

func invalidateCache() error {
	return redisClient.Del(ctx, "repos:all").Err()
}

func TEMP() { // rename back to main on prod
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
		// Try to get from cache first
		cachedRepos, err := getCachedRepos()
		if err == nil && cachedRepos != nil {
			c.JSON(200, gin.H{"data": cachedRepos, "cached": true})
			return
		}

		// Cache miss or error, fetch fresh data
		allRepos, err := getAllRepos()
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch repos"})
			return
		}

		// Try to cache the results
		if err := setCachedRepos(allRepos); err != nil {
			// Log but don't fail the request
			gin.DefaultErrorWriter.Write([]byte("Warning: failed to cache repos: " + err.Error()))
		}

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

		// Update cache with fresh data
		if err := setCachedRepos(allRepos); err != nil {
			gin.DefaultErrorWriter.Write([]byte("Warning: failed to cache repos: " + err.Error()))
		}

		c.JSON(200, gin.H{"message": "synced to django", "count": len(allRepos)})
	})

	// endpoint to check cache status
	r.GET("/api/cache/status", func(c *gin.Context) {
		cachedRepos, err := getCachedRepos()
		if err != nil {
			c.JSON(500, gin.H{"error": "cache error", "details": err.Error()})
			return
		}

		if cachedRepos == nil {
			c.JSON(200, gin.H{"status": "empty"})
			return
		}

		c.JSON(200, gin.H{"status": "cached", "count": len(cachedRepos)})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}