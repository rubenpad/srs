package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rubenpad/srs/internal/domain/entity"

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
		slog.Error("http error", "error", err)
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
	var (
		quoteErr, recErr error
		wg               sync.WaitGroup
		quote            finnhub.Quote
		recommendations  []finnhub.RecommendationTrend
	)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		data, _, err := s.finnhubClient.Quote(ctx).Symbol(ticker).Execute()
		if err != nil {
			quoteErr = err
			slog.Error("error getting stock quote", "error", err, "ticker", ticker)
			return
		}
		quote = data
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		data, _, err := s.finnhubClient.RecommendationTrends(ctx).Symbol(ticker).Execute()
		if err != nil {
			recErr = err
			slog.Error("error getting stock recommendation trends", "error", err, "ticker", ticker)
			return
		}
		recommendations = data
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			slog.Error("request timeout exceeded", "ticker", ticker)
		}
		return nil
	case <-done:
		if quoteErr != nil && recErr != nil {
			return nil
		}

		return &entity.StockDetails{
			Quote:           &quote,
			Recommendations: &recommendations,
		}
	}
}
