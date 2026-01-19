package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
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
	starRanges := []string{
		"stars:100..200",
		"stars:201..300",
		"stars:301..400",
		"stars:401..500",
		"stars:501..600",
		"stars:601..700",
		"stars:701..800",
		"stars:801..900",
		"stars:901..1000",
		"stars:1001..1100",
		"stars:1101..1200",
		"stars:1201..1300",
	}
	type ReadmeData struct {
		Readme string `json:"readme"` // struct so its not a single array and easier to loop over
	}

	var allReadmes []ReadmeData // fixxed placement so json should no longer overwrite

	for _, starRange := range starRanges {
		log.Printf("\n Querying range: %s\n", starRange)

		cursor := ""
		page := 1
		totalPages := 9 // Adjust this to liking but past 10 you get limited and need to change part of query or do stars lilke i  did(10 pages suggested or make query more specific)

		for page <= totalPages {
			log.Printf("\nFetching page %d/%d...\n", page, totalPages)

			afterClause := ""
			if cursor != "" {
				afterClause = fmt.Sprintf(`, after: "%s"`, cursor)
			}

			query := fmt.Sprintf(`
		query {
			search(type: REPOSITORY, query: "%s fork:false", first: 100%s) {
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
							... on Blob { text }
						}
						readmeLower: object(expression: "HEAD:readme.md") {
							... on Blob { text }
						}
					}
				}
			}
		}
		`, starRange, afterClause)

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
				// Checks for uppcase "README.md"
				if repo.Readme != nil && repo.Readme.Text != "" {
					readmeText = repo.Readme.Text
					readmeText = Strip(readmeText)
					readmeText = strings.ReplaceAll(readmeText, "\n", "")
					// Falls back if its lowercase, "readme.md"
				} else if repo.ReadmeLower != nil && repo.ReadmeLower.Text != "" {
					readmeText = repo.ReadmeLower.Text
					readmeText = Strip(readmeText)
					readmeText = strings.ReplaceAll(readmeText, "\n", "") // Strip left to many "\n"
				}

				if readmeText != "" {
					allReadmes = append(allReadmes, ReadmeData{Readme: readmeText})
					foundInPage++
				}
			}

			log.Printf("✅ Page %d: Found %d READMEs (Total: %d)\n", page, foundInPage, len(allReadmes))

			if !resp.Search.PageInfo.HasNextPage {
				log.Println(" No more pages available")
				break
			}

			cursor = resp.Search.PageInfo.EndCursor
			page++

			time.Sleep(1 * time.Second)
		}

		if len(allReadmes) == 0 {
			log.Fatal("No READMEs found")
		}

		log.Printf("\nSaving %d READMEs...\n", len(allReadmes))

	}
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

	log.Printf("Successfully saved %d READMEs to readme_training.json\n", len(allReadmes))
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

type Options struct {
	SkipImages bool
}

var (
	listLeadersReg = regexp.MustCompile(`(?m)^([\s\t]*)([\*\-\+]|\d\.)\s+`)

	headerReg = regexp.MustCompile(`\n={2,}`)
	strikeReg = regexp.MustCompile(`~~`)
	codeReg   = regexp.MustCompile("`{3}" + `.*\n`)

	htmlReg         = regexp.MustCompile("<[^>]+>")
	emphReg         = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	emphReg2        = regexp.MustCompile(`\*([^*]+)\*`)
	emphReg3        = regexp.MustCompile(`__([^_]+)__`)
	emphReg4        = regexp.MustCompile(`_([^_]+)_`)
	setextHeaderReg = regexp.MustCompile(`^[=\-]{2,}\s*$`)
	footnotesReg    = regexp.MustCompile(`\[\^.+?\](\: .*?$)?`)
	footnotes2Reg   = regexp.MustCompile(`\s{0,2}\[.*?\]: .*?$`)
	imagesReg       = regexp.MustCompile(`\!\[(.*?)\]\s?[\[\(].*?[\]\)]`)
	linksReg        = regexp.MustCompile(`\[(.*?)\][\[\(].*?[\]\)]`)
	blockquoteReg   = regexp.MustCompile(`>\s*`)
	refLinkReg      = regexp.MustCompile(`^\s{1,2}\[(.*?)\]: (\S+)( ".*?")?\s*$`)
	atxHeaderReg    = regexp.MustCompile(`(?m)^\#{1,6}\s*([^#]+)\s*(\#{1,6})?$`)
	atxHeaderReg2   = regexp.MustCompile(`([\*_]{1,3})(\S.*?\S)?P1`)
	atxHeaderReg3   = regexp.MustCompile("(?m)(`{3,})" + `(.*?)?P1`)
	atxHeaderReg4   = regexp.MustCompile(`^-{3,}\s*$`)
	atxHeaderReg5   = regexp.MustCompile("`(.+?)`")
)

func Strip(s string) string {
	return StripOptions(s, Options{})
}

func StripOptions(s string, opts Options) string {
	res := s
	res = listLeadersReg.ReplaceAllString(res, "$1")

	res = headerReg.ReplaceAllString(res, "\n")
	res = strikeReg.ReplaceAllString(res, "")
	res = codeReg.ReplaceAllString(res, "")

	res = emphReg.ReplaceAllString(res, "$1")
	res = emphReg2.ReplaceAllString(res, "$1")
	res = emphReg3.ReplaceAllString(res, "$1")
	res = emphReg4.ReplaceAllString(res, "$1")
	res = htmlReg.ReplaceAllString(res, "")
	res = setextHeaderReg.ReplaceAllString(res, "")
	res = footnotesReg.ReplaceAllString(res, "")
	res = footnotes2Reg.ReplaceAllString(res, "")
	if opts.SkipImages {
		res = imagesReg.ReplaceAllString(res, "")
	} else {
		res = imagesReg.ReplaceAllString(res, "$1")
	}
	res = linksReg.ReplaceAllString(res, "$1")
	res = blockquoteReg.ReplaceAllString(res, "  ")
	res = refLinkReg.ReplaceAllString(res, "")
	res = atxHeaderReg.ReplaceAllString(res, "$1")
	res = atxHeaderReg2.ReplaceAllString(res, "$2")
	res = atxHeaderReg3.ReplaceAllString(res, "$2")
	res = atxHeaderReg4.ReplaceAllString(res, "")
	res = atxHeaderReg5.ReplaceAllString(res, "$1")
	return res
}
