package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"michiru/config"
	"michiru/internal/clients"
	"michiru/models"
)

func decodeQueryParams(r *http.Request) (*models.QueryParams, error) {
	reqParams := r.URL.Query()

	query := reqParams.Get("query")
	if query == "" {
		return nil, errors.New("query cannot be empty")
	}

	limitStr := reqParams.Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, fmt.Errorf("limit must be an integer: %w", err)
	}
	if limit > 50 {
		return nil, errors.New("limit cannot exceed 50")
	}

	offsetStr := reqParams.Get("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return nil, fmt.Errorf("offset must be an integer: %w", err)
	}

	return &models.QueryParams{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func toPaging(
	req *url.URL, params *models.QueryParams, count int,
) models.PagingResponse {
	queryParams := params.ToQueryString()

	var next string
	var prev string
	resp := models.PagingResponse{Count: count}

	nextOffset := params.Offset + params.Limit
	if nextOffset < count {
		queryParams.Set("offset", strconv.Itoa(nextOffset))
		next = fmt.Sprintf("%s?%s", req.Path, queryParams.Encode())
	}

	if prevOffset := params.Offset - params.Limit; prevOffset >= 0 {
		queryParams.Set("offset", strconv.Itoa(prevOffset))
		prev = fmt.Sprintf("%s?%s", req.Path, queryParams.Encode())
	} else if params.Offset != 0 {
		queryParams.Set("offset", "0")
		prev = fmt.Sprintf("%s?%s", req.Path, queryParams.Encode())
	}

	if next != "" {
		resp.Next = &next
	}
	if prev != "" {
		resp.Prev = &prev
	}
	return resp
}

func HandleSearch(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := decodeQueryParams(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, count, err := clients.SearchAnime(cfg, params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		paging := toPaging(r.URL, params, count)

		resp := models.QueryResponse{
			Payload: data,
			Paging:  paging,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		return
	}
}

func HandleMetadata(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		meta, err := clients.GetMetadata(r.Context(), cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// The importer runs daily by default, so data should be valid for a day
		const validDuration = 24 * time.Hour
		expiresTime := meta.RetrievedAt.Add(validDuration)

		var maxAgeSeconds int
		if time.Now().Before(expiresTime) {
			maxAgeSeconds = int(time.Until(expiresTime).Seconds())
		} else {
			maxAgeSeconds = 0
		}

		// Set headers to support caching, revalidate every 24 hours
		w.Header().Set(
			"Cache-Control", fmt.Sprintf("public, max-age=%d", maxAgeSeconds),
		)
		w.Header().Set("Expires", expiresTime.UTC().Format(http.TimeFormat))
		w.Header().Set(
			"Last-Modified", meta.RetrievedAt.UTC().Format(http.TimeFormat),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(meta)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}
}
