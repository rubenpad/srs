package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

const errorMessage = "there was an error while processing the stock ratings from external API"

type finnhubResponse struct {
	Quote           *finnhub.Quote                 `json:"quote"`
	Recommendations *[]finnhub.RecommendationTrend `json:"recommendations"`
}

type stockRatingsDto struct {
	NextPage string               `json:"next_page"`
	Items    []entity.StockRating `json:"items"`
}

type StockRatingApi struct {
	finnhubClient *finnhub.DefaultApiService
	httpClient    *http.Client
	baseURL       string
	authToken     string
}

func NewStockRatingApi() *StockRatingApi {
	configuration := finnhub.NewConfiguration()
	configuration.AddDefaultHeader("X-Finnhub-Token", os.Getenv("FINNHUB_API_KEY"))

	return &StockRatingApi{
		httpClient:    &http.Client{},
		baseURL:       os.Getenv("STOCK_RATING_API_URL"),
		authToken:     os.Getenv("STOCK_RATING_API_AUTH_TOKEN"),
		finnhubClient: finnhub.NewAPIClient(configuration).DefaultApi,
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

func (s *StockRatingApi) GetStockDetails(ctx context.Context, ticker string) *finnhubResponse {
	type result[T any] struct {
		data T
		err  error
	}

	quoteCh := make(chan result[*finnhub.Quote])
	recommendationCh := make(chan result[*[]finnhub.RecommendationTrend])

	go func() {
		data, _, err := s.finnhubClient.Quote(context.Background()).Symbol(ticker).Execute()
		quoteCh <- result[*finnhub.Quote]{&data, err}
	}()

	go func() {
		data, _, err := s.finnhubClient.RecommendationTrends(context.Background()).Symbol(ticker).Execute()
		recommendationCh <- result[*[]finnhub.RecommendationTrend]{&data, err}
	}()

	var quote *finnhub.Quote
	var recommendations *[]finnhub.RecommendationTrend
	for range 2 {
		select {
		case res := <-quoteCh:
			if res.err != nil {
				slog.Error("error getting stock quote", "error", res.err)
				return nil
			}
			quote = res.data

		case res := <-recommendationCh:
			if res.err != nil {
				slog.Error("error getting stock recommendation trends", "error", res.err)
				return nil
			}
			recommendations = res.data
		}
	}

	return &finnhubResponse{
		Quote:           quote,
		Recommendations: recommendations,
	}
}
