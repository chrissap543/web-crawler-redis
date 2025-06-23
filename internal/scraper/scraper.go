package scraper

import(
	"net/http"
	"io"
	"regexp"
	"net/url"
	"os"
	"fmt"
	"time"
	"log"
	"strings"

	"web-crawler/internal/myredis"
)

func getHttp(q *myredis.RedisQueue, url string) string {
	resp, err := http.Get(url)
	if err != nil {
		q.Enqueue(url)
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		q.Enqueue(url)
		return ""
	}
	return string(body)
}

func getLinks(q *myredis.RedisQueue, baseURL string) []string {
	html := getHttp(q, baseURL)

	re := regexp.MustCompile(`href\s*=\s*["']([^"']+)["']`)
    matches := re.FindAllStringSubmatch(html, -1)

    var hrefs []string
    for _, match := range matches {
        if len(match) > 1 {
            hrefs = append(hrefs, match[1])
        }
    }

    return cleanHrefs(hrefs, baseURL)
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

func ScrapePage(q *myredis.RedisQueue, s *myredis.RedisSet) {
	if err := os.MkdirAll("/app/data", 0755); err != nil {
		log.Printf("Could not open directory")
        return
    }

    // Open file for writing scraped URLs
    file, err := os.OpenFile("/app/data/scraped_urls.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        log.Printf("Error opening file: %v", err)
        return
    }
    defer file.Close()

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

		logLine := fmt.Sprintf("Scraping: %s\n", website)
		if _, err := file.WriteString(logLine); err != nil {
            log.Printf("Error writing to file: %v", err)
        }

		links := getLinks(q, website)
		for _, link := range links {
			q.Enqueue(link)
			s.Add(website)
		}

		file.Sync()
	}
}
