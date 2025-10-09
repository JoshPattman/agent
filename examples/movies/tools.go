package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/JoshPattman/agent"
)

var _ agent.Tool = &timeTool{}

type timeTool struct {
}

func (t *timeTool) Call(map[string]any) (string, error) {
	return time.Now().Format(time.ANSIC), nil
}

func (t *timeTool) Name() string {
	return "get_time"
}

func (t *timeTool) Description() []string {
	return []string{
		"Gets the current time",
		"Takes no arguments",
	}
}

type movieLookupTool struct {
	APIKey string
}

func (t *movieLookupTool) Name() string {
	return "movie_lookup"
}

func (t *movieLookupTool) Description() []string {
	return []string{
		"Searches The Movie Database (TMDb) for a movie by title.",
		"Argument: query (string) — the movie title to search for.",
	}
}

func (t *movieLookupTool) Call(args map[string]any) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("missing query argument")
	}

	if t.APIKey == "" {
		return "", fmt.Errorf("API key not set in tool")
	}

	endpoint := fmt.Sprintf(
		"https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s",
		t.APIKey,
		url.QueryEscape(query),
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("TMDb returned status %d", resp.StatusCode)
	}

	var data struct {
		Results []struct {
			Title       string  `json:"title"`
			ReleaseDate string  `json:"release_date"`
			VoteAverage float64 `json:"vote_average"`
			Overview    string  `json:"overview"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Results) == 0 {
		return "No movies found", nil
	}

	if len(data.Results) > 10 {
		data.Results = data.Results[:10]
	}
	results := make([]string, len(data.Results))
	for i, m := range data.Results {
		results[i] = fmt.Sprintf("%s (%s) — Rating %.1f/10\n%s",
			m.Title,
			m.ReleaseDate,
			m.VoteAverage,
			m.Overview,
		)
	}
	return strings.Join(results, "\n\n"), nil
}

var _ agent.Tool = &movieKeywordTool{}

type movieKeywordTool struct {
	APIKey string
}

func (t *movieKeywordTool) Name() string {
	return "movie_keyword_search"
}

func (t *movieKeywordTool) Description() []string {
	return []string{
		"Searches The Movie Database (TMDb) for movies based on keywords.",
		"Argument: query (string) — the keyword or concept to search for (e.g., 'sci fi', 'time travel').",
	}
}

func (t *movieKeywordTool) Call(args map[string]any) (string, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("missing query argument")
	}
	if t.APIKey == "" {
		return "", fmt.Errorf("API key not set in tool")
	}

	// Step 1: Search for keyword IDs
	keywordEndpoint := fmt.Sprintf(
		"https://api.themoviedb.org/3/search/keyword?api_key=%s&query=%s",
		t.APIKey,
		url.QueryEscape(query),
	)

	resp, err := http.Get(keywordEndpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("TMDb keyword search returned status %d", resp.StatusCode)
	}

	var keywordData struct {
		Results []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&keywordData); err != nil {
		return "", err
	}

	if len(keywordData.Results) == 0 {
		return "No matching keywords found", nil
	}

	// Collect keyword IDs (use top 3 for discovery)
	var keywordIDs []string
	for i, k := range keywordData.Results {
		if i >= 3 {
			break
		}
		keywordIDs = append(keywordIDs, fmt.Sprintf("%d", k.ID))
	}

	// Step 2: Discover movies with those keywords
	discoverEndpoint := fmt.Sprintf(
		"https://api.themoviedb.org/3/discover/movie?api_key=%s&with_keywords=%s",
		t.APIKey,
		strings.Join(keywordIDs, ","),
	)

	resp2, err := http.Get(discoverEndpoint)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("TMDb discover returned status %d", resp2.StatusCode)
	}

	var discoverData struct {
		Results []struct {
			Title       string  `json:"title"`
			ReleaseDate string  `json:"release_date"`
			VoteAverage float64 `json:"vote_average"`
			Overview    string  `json:"overview"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&discoverData); err != nil {
		return "", err
	}

	if len(discoverData.Results) == 0 {
		return "No movies found for these keywords", nil
	}

	if len(discoverData.Results) > 10 {
		discoverData.Results = discoverData.Results[:10]
	}

	results := make([]string, len(discoverData.Results))
	for i, m := range discoverData.Results {
		results[i] = fmt.Sprintf("%s (%s) — Rating %.1f/10\n%s",
			m.Title,
			m.ReleaseDate,
			m.VoteAverage,
			m.Overview,
		)
	}

	return strings.Join(results, "\n\n"), nil
}
