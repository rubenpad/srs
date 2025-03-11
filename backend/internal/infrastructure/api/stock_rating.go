package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rubenpad/stock-rating-system/internal/domain/entity"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

const errorMessage = "there was an error while processing the stock ratings from external API"

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

func (s *StockRatingApi) GetStockDetails(ctx context.Context, ticker string) *entity.StockDetails {
	type result[T any] struct {
		data T
		err  error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quoteCh := make(chan result[finnhub.Quote])
	recommendationCh := make(chan result[[]finnhub.RecommendationTrend])

	defer close(quoteCh)
	defer close(recommendationCh)

	go func() {
		data, _, err := s.finnhubClient.Quote(ctx).Symbol(ticker).Execute()
		select {
		case <-ctx.Done():
			return
		case quoteCh <- result[finnhub.Quote]{data, err}:
		}
	}()

	go func() {
		data, _, err := s.finnhubClient.RecommendationTrends(ctx).Symbol(ticker).Execute()
		select {
		case <-ctx.Done():
			return
		case recommendationCh <- result[[]finnhub.RecommendationTrend]{data, err}:
		}
	}()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	var quote finnhub.Quote
	var recommendations []finnhub.RecommendationTrend
	for range 2 {
		select {
		case <-ctx.Done():
			slog.Error("context cancelled", "error", ctx.Err())
			return nil

		case <-timer.C:
			slog.Error("timeout getting stock details")
			return nil

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

	return &entity.StockDetails{
		Quote:           &quote,
		Recommendations: &recommendations,
	}
}
