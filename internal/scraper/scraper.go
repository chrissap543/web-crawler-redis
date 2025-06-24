package scraper

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"web-crawler/internal/myredis"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func createPageNode(driver neo4j.DriverWithContext, ctx context.Context, url, title string) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query := `
        MERGE (p:Page {url: $url})
        ON CREATE SET
            p.title = $title,
            p.created_at = datetime(),
            p.scraped_count = 1
        ON MATCH SET
            p.scraped_count = p.scraped_count + 1,
            p.last_scraped = datetime()
        RETURN p.url
    `
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, query, map[string]any {
				"url": url,
				"title": title,
		})
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}
		return nil, result.Err()
	})
	return err
}

func createPageLink(driver neo4j.DriverWithContext, ctx context.Context, fromURL, toURL string) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    query := `
        MERGE (from:Page {url: $fromURL})
        MERGE (to:Page {url: $toURL})
        MERGE (from)-[r:LINKS_TO]->(to)
        ON CREATE SET r.created_at = datetime()
        RETURN r
    `

    _, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        result, err := tx.Run(ctx, query, map[string]any{
            "fromURL": fromURL,
            "toURL":   toURL,
        })
        if err != nil {
            return nil, err
        }

        if result.Next(ctx) {
            return result.Record().Values[0], nil
        }
        return nil, result.Err()
    })

    return err
}

func extractTitle(html string) string {
    re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
    matches := re.FindStringSubmatch(html)
    if len(matches) > 1 {
        return strings.TrimSpace(matches[1])
    }
    return "No Title"
}

func getHttpWithTitle(q *myredis.RedisQueue, url string) (string, string) {
	resp, err := http.Get(url)
	if err != nil {
		q.Enqueue(url)
		return "", ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		q.Enqueue(url)
		return "", ""
	}
	html := string(body)
	return html, extractTitle(html)
}

func getLinks(q *myredis.RedisQueue, baseURL string) ([]string, string) {
	html, title := getHttpWithTitle(q, baseURL)

	re := regexp.MustCompile(`href\s*=\s*["']([^"']+)["']`)
    matches := re.FindAllStringSubmatch(html, -1)

    var hrefs []string
    for _, match := range matches {
        if len(match) > 1 {
            hrefs = append(hrefs, match[1])
        }
    }

    return cleanHrefs(hrefs, baseURL), title
}

func cleanHrefs(hrefs []string, baseURL string) []string {
    base, err := url.Parse(baseURL)
    if err != nil {
        return []string{}
    }

    var cleanedHrefs []string
    seen := make(map[string]bool)

    for _, href := range hrefs {
        href = strings.TrimSpace(href)

        // Skip unwanted links
        if href == "" || href == "#" ||
           strings.HasPrefix(href, "javascript:") ||
           strings.HasPrefix(href, "mailto:") ||
           strings.HasPrefix(href, "tel:") {
            continue
        }

        // Parse the href
        linkURL, err := url.Parse(href)
        if err != nil {
            continue
        }

        // Make it absolute using ResolveReference
        absoluteURL := base.ResolveReference(linkURL)
		if (/*absoluteURL.Host != "en.wikipedia.org" && */ absoluteURL.Host != "simple.wikipedia.org") ||
			!strings.HasPrefix(absoluteURL.Path, "/wiki/") || strings.Contains(absoluteURL.Path, ":") {
			continue
		}

        absoluteStr := absoluteURL.String()
        // Remove duplicates
        if !seen[absoluteStr] {
            seen[absoluteStr] = true
            cleanedHrefs = append(cleanedHrefs, absoluteStr)
        }
    }

    return cleanedHrefs

}

func ScrapePage(q *myredis.RedisQueue, s *myredis.RedisSet, driver neo4j.DriverWithContext) {
	// if err := os.MkdirAll("/app/data", 0755); err != nil {
	// 	log.Printf("Could not open directory")
    // }

    // Open file for writing scraped URLs
    // file, err := os.OpenFile("/app/data/scraped_urls.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    // if err != nil {
    //     log.Printf("Error opening file: %v", err)
    // } else {
	// 	defer file.Close()
	// }

	ctx := context.Background()

	for {
		website, err := q.Dequeue(time.Second * 5)
		if err != nil {
			continue
		}

		isScraped, err := s.IsMember(website)
		if err != nil {
			continue
		}
		if isScraped {
			log.Printf("Already scraped: %s\n", website)
			continue
		}

		links, title := getLinks(q, website)

		if driver != nil {
			if err := createPageNode(driver, ctx, website, title); err != nil {
				log.Printf("Error creating page node for %s: %v", website, err)
			}
		}

		linkCount := 0
		for _, link := range links {
			q.Enqueue(link)

			if driver != nil {
				if err := createPageLink(driver, ctx, website, link); err != nil {
					log.Printf("Error creating link: %v", err)
				} else {
					linkCount++
				}
			}

		}
		s.Add(website)
		// file.Sync()
	}
}
