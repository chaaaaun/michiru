package models

import (
	"net/url"
	"strconv"
)

// API request params

type QueryParams struct {
	Query  string
	Limit  int
	Offset int
}

func (q *QueryParams) ToQueryString() url.Values {
	return url.Values{
		"query":  {q.Query},
		"limit":  {strconv.Itoa(q.Limit)},
		"offset": {strconv.Itoa(q.Offset)},
	}
}

type AnimeSearchDocument struct {
	AnimeDocument
	Formatted       AnimeDocument          `json:"_formatted"`
	MatchesPosition map[string]interface{} `json:"_matchesPosition,omitempty"`
	RankingScore    float64                `json:"_rankingScore,omitempty"`
}

// JSON response structs

type QueryResponse struct {
	Payload []AnimeSearchDocument `json:"payload"`
	Paging  PagingResponse        `json:"paging"`
}

type PagingResponse struct {
	Count int     `json:"count"`
	Next  *string `json:"next,omitempty"`
	Prev  *string `json:"prev,omitempty"`
}
