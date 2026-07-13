package main

import (
	"database/sql"
	"net/http"
	"time"
)

type searchCatalogProvider struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	CostPerQuery     float64  `json:"costPerQuery"`
	FreeMonthlyQuota int      `json:"freeMonthlyQuota"`
	SearchTypes      []string `json:"searchTypes,omitempty"`
	FetchFormats     []string `json:"fetchFormats,omitempty"`
	Status           string   `json:"status"`
	ConfigureHref    string   `json:"configureHref"`
	NoAuth           bool     `json:"-"`
}

var searchCatalog = []searchCatalogProvider{
	{ID: "serper-search", Name: "Serper Search", Kind: "search", CostPerQuery: .001, FreeMonthlyQuota: 2500, SearchTypes: []string{"web", "news"}},
	{ID: "brave-search", Name: "Brave Search", Kind: "search", CostPerQuery: .005, FreeMonthlyQuota: 1000, SearchTypes: []string{"web", "news"}},
	{ID: "perplexity-search", Name: "Perplexity Search", Kind: "search", CostPerQuery: .005, SearchTypes: []string{"web"}},
	{ID: "exa-search", Name: "Exa Search", Kind: "search", CostPerQuery: .007, FreeMonthlyQuota: 1000, SearchTypes: []string{"web", "news"}},
	{ID: "tavily-search", Name: "Tavily Search", Kind: "search", CostPerQuery: .008, FreeMonthlyQuota: 1000, SearchTypes: []string{"web", "news"}},
	{ID: "google-pse-search", Name: "Google Programmable Search", Kind: "search", CostPerQuery: .005, FreeMonthlyQuota: 3000, SearchTypes: []string{"web", "news"}},
	{ID: "linkup-search", Name: "Linkup Search", Kind: "search", CostPerQuery: .005, FreeMonthlyQuota: 1000, SearchTypes: []string{"web"}},
	{ID: "searchapi-search", Name: "SearchAPI", Kind: "search", CostPerQuery: .004, FreeMonthlyQuota: 100, SearchTypes: []string{"web", "news"}},
	{ID: "youcom-search", Name: "You.com Search", Kind: "search", CostPerQuery: .005, SearchTypes: []string{"web", "news"}},
	{ID: "searxng-search", Name: "SearXNG Search", Kind: "search", FreeMonthlyQuota: 999999, SearchTypes: []string{"web", "news"}, NoAuth: true},
	{ID: "ollama-search", Name: "Ollama Search", Kind: "search", FreeMonthlyQuota: 1000, SearchTypes: []string{"web"}},
	{ID: "zai-search", Name: "Z.AI Coding Plan Search", Kind: "search", SearchTypes: []string{"web"}},
	{ID: "duckduckgo-search", Name: "DuckDuckGo Search", Kind: "search", FreeMonthlyQuota: 999999, SearchTypes: []string{"web"}, NoAuth: true},
	{ID: "firecrawl", Name: "Firecrawl", Kind: "fetch", CostPerQuery: .002, FreeMonthlyQuota: 500, FetchFormats: []string{"markdown", "html", "links", "screenshot"}},
	{ID: "jina-reader", Name: "Jina Reader", Kind: "fetch", CostPerQuery: .0005, FreeMonthlyQuota: 1000, FetchFormats: []string{"markdown", "text"}},
	{ID: "tavily-search", Name: "Tavily Extract", Kind: "fetch", CostPerQuery: .001, FreeMonthlyQuota: 1000, FetchFormats: []string{"markdown", "text"}},
	{ID: "tinyfish", Name: "TinyFish Fetch", Kind: "fetch", FetchFormats: []string{"markdown", "html"}},
}

func searchProvidersHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configured := map[string]bool{}
		rows, err := dbConn.Query("SELECT provider FROM provider_connections WHERE is_active=1 AND (api_key!='' OR access_token!='')")
		if err != nil {
			jsonError(w, http.StatusInternalServerError, "Failed to list providers")
			return
		}
		for rows.Next() {
			var provider string
			if err := rows.Scan(&provider); err != nil {
				rows.Close()
				jsonError(w, http.StatusInternalServerError, "Failed to list providers")
				return
			}
			configured[provider] = true
		}
		rows.Close()
		providers := make([]searchCatalogProvider, len(searchCatalog))
		copy(providers, searchCatalog)
		created := time.Now().Unix()
		legacy := make([]map[string]interface{}, 0, len(providers))
		for i := range providers {
			providers[i].ConfigureHref = "/dashboard/providers"
			providers[i].Status = "missing"
			if providers[i].NoAuth || configured[providers[i].ID] || providers[i].ID == "perplexity-search" && configured["perplexity"] {
				providers[i].Status = "configured"
			}
			legacy = append(legacy, map[string]interface{}{"id": providers[i].ID, "object": "search_provider", "created": created, "name": providers[i].Name, "search_types": providers[i].SearchTypes})
		}
		writeJSONResponse(w, map[string]interface{}{"providers": providers, "data": legacy})
	}
}
