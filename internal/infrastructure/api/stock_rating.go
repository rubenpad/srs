package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"
)

const errorMessage = "there was an error while processing the stock ratings from external API"

type stockRatingsDto struct {
	NextPage string               `json:"next_page"`
	Items    []entity.StockRating `json:"items"`
}

type StockRatingApi struct {
	httpClient *http.Client
	baseURL    string
	authToken  string
}

func NewStockRatingApi() *StockRatingApi {
	return &StockRatingApi{
		httpClient: &http.Client{},
		baseURL:    os.Getenv("STOCK_RATING_API_URL"),
		authToken:  os.Getenv("STOCK_RATING_API_AUTH_TOKEN"),
	}
}

func (s *StockRatingApi) GetStockRatings(ctx context.Context, nextPage string) ([]entity.StockRating, string, error) {
	baseURL := s.baseURL + "/swechallenge/list"

	url := baseURL
	if nextPage != "" {
		url = fmt.Sprintf("%s?next_page=%s", baseURL, nextPage)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nextPage, errors.New(errorMessage)
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.authToken))
	response, err := s.httpClient.Do(request)
	if err != nil {
		return nil, nextPage, errors.New(errorMessage)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, nextPage, errors.New(errorMessage)
	}

	var stockRatings stockRatingsDto
	if err := json.NewDecoder(response.Body).Decode(&stockRatings); err != nil {
		return nil, nextPage, errors.New(errorMessage)
	}

	return stockRatings.Items, stockRatings.NextPage, nil
}
