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

	"github.com/gocolly/colly/v2"
	"github.com/rubenpad/srs/internal/domain/entity"

	"github.com/cenkalti/backoff/v5"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

const errorMessage = "there was an error while processing the stock ratings from external API"

type stockRatingsDto struct {
	NextPage string               `json:"next_page"`
	Items    []entity.StockRating `json:"items"`
}

type StockRatingApi struct {
	baseURL       string
	authToken     string
	httpClient    *http.Client
	collector     colly.Collector
	finnhubClient *finnhub.DefaultApiService
}

func NewStockRatingApi() *StockRatingApi {
	configuration := finnhub.NewConfiguration()
	configuration.AddDefaultHeader("X-Finnhub-Token", os.Getenv("FINNHUB_API_KEY"))

	return &StockRatingApi{
		httpClient:    &http.Client{},
		baseURL:       os.Getenv("STOCK_RATING_API_URL"),
		authToken:     os.Getenv("STOCK_RATING_API_AUTH_TOKEN"),
		finnhubClient: finnhub.NewAPIClient(configuration).DefaultApi,
		collector:     *colly.NewCollector(),
	}
}

func (s *StockRatingApi) GetStockRatings(ctx context.Context, nextPage string) ([]entity.StockRating, string, error) {
	baseURL := s.baseURL + "/swechallenge/list"

	url := baseURL
	if nextPage != "" {
		url = fmt.Sprintf("%s?next_page=%s", baseURL, nextPage)
	}

	operation := func() ([]entity.StockRating, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			slog.Error("http error", "error", err)
			return nil, errors.New(errorMessage)
		}

		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.authToken))
		response, err := s.httpClient.Do(request)
		if err != nil {
			return nil, errors.New(errorMessage)
		}

		defer response.Body.Close()

		if response.StatusCode == http.StatusOK {
			var stockRatings stockRatingsDto
			if err := json.NewDecoder(response.Body).Decode(&stockRatings); err != nil {
				slog.Error("error decoding stock ratings", "error", err)
				return nil, errors.New(errorMessage)
			}

			nextPage = stockRatings.NextPage
			return stockRatings.Items, nil
		}

		if response.StatusCode >= 400 && response.StatusCode <= 499 {
			slog.Error("client error from external API", "status", response.StatusCode)
			return nil, backoff.Permanent(errors.New(errorMessage))
		}

		slog.Info("error getting stock ratings - will retry", "status", response.StatusCode)
		return nil, errors.New(errorMessage)
	}

	result, err := backoff.Retry(
		ctx,
		operation,
		backoff.WithMaxTries(3),
		backoff.WithMaxElapsedTime(1*time.Minute),
		backoff.WithBackOff(backoff.NewExponentialBackOff()))

	return result, nextPage, err
}

func (s *StockRatingApi) GetStockDetails(ctx context.Context, ticker string) *entity.StockDetails {
	var (
		quoteErr, recErr error
		wg               sync.WaitGroup

		keyFacts        string
		quote           finnhub.Quote
		recommendations []finnhub.RecommendationTrend
	)

	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	tickerUrl := fmt.Sprintf("%s/%s", os.Getenv("WEB_TICKER_DATA_URL"), ticker)

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.collector.OnError(func(r *colly.Response, err error) {
			slog.Error("error collecting information from web", "error", err)
		})

		s.collector.OnHTML("html > body > div:nth-of-type(1) > section:nth-of-type(3) > section:nth-of-type(1) > div > section > div > div:nth-of-type(2)", func(e *colly.HTMLElement) {
			keyFacts = e.ChildText("p")
		})

		s.collector.OnRequest(func(r *colly.Request) {
			slog.Info("visiting web", "url", tickerUrl)
		})

		s.collector.Visit(tickerUrl)
	}()

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
			KeyFacts:        keyFacts,
			Quote:           &quote,
			Recommendations: &recommendations,
		}
	}
}
