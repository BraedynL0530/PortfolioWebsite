package main

import (
	"encoding/json"
	//"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/get-repos-names", func(c *gin.Context) {
		resp, err := http.Get("https://api.github.com/users/BraedynL0530/repos")
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch repos"})
		}
		defer resp.Body.Close()

		var nameData interface{}
		json.NewDecoder(resp.Body).Decode(&nameData)

		c.JSON(200, nameData)
		repos, ok := nameData.([]interface{})
		if !ok {
			c.JSON(500, gin.H{"error": "unexpected data format"})
			return
		}

		var repoNames []string
		for _, r := range repos {
			repoMap := r.(map[string]interface{})
			repoNames = append(repoNames, repoMap["name"].(string))
		}

	})

	r.GET("/get-readme/:repoName", func(c *gin.Context) {
		repoName := c.Param("repoName") // fetch repoName from URL

		resp, err := http.Get("https://api.github.com/repos/BraedynL0530/" + repoName + "/readme")
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch README"})
			return
		}
		defer resp.Body.Close()

		var readmeData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&readmeData); err != nil {
			c.JSON(500, gin.H{"error": "failed to decode README"})
			return
		}

		c.JSON(200, readmeData)
	})
	r.POST("/api/")

}
