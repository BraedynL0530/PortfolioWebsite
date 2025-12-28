package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/machinebox/graphql"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}
	client := graphql.NewClient("https://api.github.com/graphql")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN not set")
	}

	var allReadmes []string
	cursor := ""
	page := 1
	totalPages := 10 // Adjust this to get more (10 pages = ~1000 repos)

	for page <= totalPages {
		log.Printf("\n📄 Fetching page %d/%d...\n", page, totalPages)

		afterClause := ""
		if cursor != "" {
			afterClause = fmt.Sprintf(`, after: "%s"`, cursor)
		}

		query := fmt.Sprintf(`
        query {
            search(type: REPOSITORY, query: "stars:>100 fork:false", first: 100%s) {
                pageInfo {
                    hasNextPage
                    endCursor
                }
                nodes {
                    ... on Repository {
                        nameWithOwner
                        stargazerCount
                        primaryLanguage {
                            name
                        }
                        readme: object(expression: "HEAD:README.md") {
                            ... on Blob {
                                text
                            }
                        }
                        readmeLower: object(expression: "HEAD:readme.md") {
                            ... on Blob {
                                text
                            }
                        }
                    }
                }
            }
        }
        `, afterClause)

		request := graphql.NewRequest(query)
		request.Header.Set("Authorization", "Bearer "+token)

		var resp GithubReposData
		if err := client.Run(context.Background(), request, &resp); err != nil {
			log.Printf("❌ Error on page %d: %v\n", page, err)
			break
		}

		foundInPage := 0
		for _, repo := range resp.Search.Nodes {
			var readmeText string

			if repo.Readme != nil && repo.Readme.Text != "" {
				readmeText = repo.Readme.Text
			} else if repo.ReadmeLower != nil && repo.ReadmeLower.Text != "" {
				readmeText = repo.ReadmeLower.Text
			}

			if readmeText != "" {
				allReadmes = append(allReadmes, readmeText)
				foundInPage++
			}
		}

		log.Printf("✅ Page %d: Found %d READMEs (Total: %d)\n", page, foundInPage, len(allReadmes))

		if !resp.Search.PageInfo.HasNextPage {
			log.Println("📍 No more pages available")
			break
		}

		cursor = resp.Search.PageInfo.EndCursor
		page++

		// Be nice to GitHub API - rate limiting
		time.Sleep(1 * time.Second)
	}

	if len(allReadmes) == 0 {
		log.Fatal("❌ No READMEs found")
	}

	log.Printf("\n💾 Saving %d READMEs...\n", len(allReadmes))

	file, err := os.Create("readme_training.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(allReadmes); err != nil {
		log.Fatal(err)
	}

	log.Printf("🎉 Successfully saved %d READMEs to readme_training.json\n", len(allReadmes))
}

type GithubReposData struct {
	Search struct {
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
		Nodes []struct {
			NameWithOwner   string `json:"nameWithOwner"`
			StargazerCount  int    `json:"stargazerCount"`
			PrimaryLanguage *struct {
				Name string `json:"name"`
			} `json:"primaryLanguage"`
			Readme *struct {
				Text string `json:"text"`
			} `json:"readme"`
			ReadmeLower *struct {
				Text string `json:"text"`
			} `json:"readmeLower"`
		} `json:"nodes"`
	} `json:"search"`
}
