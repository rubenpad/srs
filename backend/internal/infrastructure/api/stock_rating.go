package api

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gocolly/colly/v2"
	"github.com/rubenpad/srs/internal/domain/entity"

	"github.com/cenkalti/backoff/v5"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

const errorMessage = "there was an error while processing the stock ratings from external API"

var actions = []string{
	"target lowered by",
	"target raised by",
	"reiterated by",
	"downgraded by",
	"targetset by",
	"initiated by",
	"upgraded by",
}

var ratings = []string{
	"Sector Underperform",
	"Sector Outperform",
	"Market Outperform",
	"Market Perform",
	"Sector Perform",
	"Sector Weight",
	"Equal Weight",
	"Peer Perform",
	"Outperformer",
	"Underperform",
	"Underweight",
	"Overweight",
	"Strong-Buy",
	"Outperform",
	"Top Pick",
	"Positive",
	"Negative",
	"Neutral",
	"In-Line",
	"Reduce",
	"Inline",
	"Buy",
	"Sell",
	"Hold",
}

type stockRatingsDto struct {
	NextPage string               `json:"next_page"`
	Items    []entity.StockRating `json:"items"`
}

type StockRatingApi struct {
	baseURL       string
	authToken     string
	format        string
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
		format:        os.Getenv("STOCK_RATING_API_FORMAT"),
		finnhubClient: finnhub.NewAPIClient(configuration).DefaultApi,
		collector:     *colly.NewCollector(),
	}
}

func (s *StockRatingApi) GetStockRatings(ctx context.Context, nextPage string, useCustomFormat bool) ([]entity.StockRating, string, error) {
	url := s.baseURL + "/swechallenge/list"
	withCustomFormat := useCustomFormat && s.format != ""

	operation := func() ([]entity.StockRating, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		q := request.URL.Query()

		if withCustomFormat {
			q.Set("format", s.format)
		}

		if nextPage != "" {
			q.Set("next_page", nextPage)
		}

		request.URL.RawQuery = q.Encode()

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
			if withCustomFormat {
				reader := bufio.NewReader(response.Body)
				firstLine, readErr := reader.ReadString('\n')

				if readErr != nil && readErr != io.EOF {
					slog.Error("error reading first line for next page", "error", readErr)
					return nil, backoff.Permanent(errors.New(errorMessage))
				}

				nextPage = strings.TrimSpace(firstLine)

				ratings, parseErr := s.parseStockRatingsResponse(reader)
				if parseErr != nil {
					return nil, backoff.Permanent(parseErr)
				}

				return ratings, nil
			}

			var stockRatings stockRatingsDto
			if err := json.NewDecoder(response.Body).Decode(&stockRatings); err != nil {
				slog.Error("error decoding stock ratings", "error", err)
				return nil, backoff.Permanent(errors.New(errorMessage))
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

func (s *StockRatingApi) parseStockRatingsResponse(body io.Reader) ([]entity.StockRating, error) {
	scanner := bufio.NewScanner(body)

	var stockRatings []entity.StockRating
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "$") {
			continue
		}

		rating, err := parseStockRatingLine(line)
		if err != nil {
			slog.Error("error parsing stock rating line", "error", err, "line", line)
			continue
		}

		stockRatings = append(stockRatings, rating)
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error scanning response body", "error", err)
		return nil, fmt.Errorf("%s: %w", errorMessage, err)
	}

	return stockRatings, nil
}

func parseStockRatingLine(line string) (entity.StockRating, error) {
	dateRegex := regexp.MustCompile(`((Fri|Mon|Tue|Wed|Thu|Sat|Sun)[A-Za-z]{3}\d{6}\d{2}:\d{2}UTC)`)
	dateMatch := dateRegex.FindString(line)
	if dateMatch == "" {
		return entity.StockRating{}, fmt.Errorf("invalid stock rating format: date not found in '%s'", line)
	}

	dateIndex := strings.Index(line, dateMatch)
	if dateIndex == -1 {
		return entity.StockRating{}, fmt.Errorf("internal error: date index not found in '%s'", line)
	}

	dataPart := line[:dateIndex]
	partsRegex := regexp.MustCompile(`^([A-Z]+)(\$\d+(?:,\d{3})*\.\d{2})(\$\d+(?:,\d{3})*\.\d{2})(.*)$`)
	parts := partsRegex.FindStringSubmatch(dataPart)

	if len(parts) != 5 {
		return entity.StockRating{}, fmt.Errorf("invalid stock rating format: ticker/targets not matched in '%s'", dataPart)
	}

	ticker := parts[1]
	targetFrom := parts[2]
	targetTo := parts[3]
	remainingPart := parts[4]

	var action, companyPart, brokerageRatingsPart string

	for _, knownAction := range actions {
		normalizedAction := strings.ReplaceAll(knownAction, " ", "")
		index := strings.Index(remainingPart, normalizedAction)
		if index != -1 {
			action = knownAction
			companyPart = remainingPart[:index]
			brokerageRatingsPart = remainingPart[index+len(normalizedAction):]
			break
		}
	}

	if action == "" {
		return entity.StockRating{}, fmt.Errorf("action not found in '%s'", remainingPart)
	}

	var ratingFrom, ratingTo string

	tempBrokerageRatingsPart := brokerageRatingsPart

	for _, knownRating := range ratings {
		normalizedRating := strings.ReplaceAll(knownRating, " ", "")
		if strings.HasSuffix(tempBrokerageRatingsPart, normalizedRating) {
			ratingTo = knownRating
			tempBrokerageRatingsPart = strings.TrimSuffix(tempBrokerageRatingsPart, normalizedRating)
			break
		}
	}

	if ratingTo == "" {
		return entity.StockRating{}, fmt.Errorf("ratingTo not found in '%s'", brokerageRatingsPart)
	}

	for _, knownRating := range ratings {
		normalizedRating := strings.ReplaceAll(knownRating, " ", "")
		if strings.HasSuffix(tempBrokerageRatingsPart, normalizedRating) {
			ratingFrom = knownRating
			tempBrokerageRatingsPart = strings.TrimSuffix(tempBrokerageRatingsPart, normalizedRating)
			break
		}
	}

	if ratingFrom == "" {
		return entity.StockRating{}, fmt.Errorf("ratingFrom not found in '%s'", brokerageRatingsPart)
	}

	parsedTime, err := time.Parse("MonJan02200615:04MST", dateMatch)
	if err != nil {
		return entity.StockRating{}, fmt.Errorf("error parsing date '%s': %w", dateMatch, err)
	}

	return entity.StockRating{
		Brokerage:  formatConcatenatedString(tempBrokerageRatingsPart),
		Action:     action,
		Company:    formatConcatenatedString(companyPart),
		Ticker:     ticker,
		RatingFrom: ratingFrom,
		RatingTo:   ratingTo,
		TargetFrom: targetFrom,
		TargetTo:   targetTo,
		Time:       parsedTime.Truncate(24 * time.Hour),
	}, nil
}

func formatConcatenatedString(name string) string {
	if name == "" {
		return ""
	}

	var builder strings.Builder
	runes := []rune(name)

	for i, r := range runes {
		builder.WriteRune(r)

		if i < len(runes)-1 {
			nextR := runes[i+1]

			if (unicode.IsLower(r) || unicode.IsDigit(r)) && unicode.IsUpper(nextR) {
				builder.WriteRune(' ')
				continue
			}

			if (unicode.IsLetter(r) || unicode.IsDigit(r) || r == ')') && (nextR == '(' || nextR == '&') {
				builder.WriteRune(' ')
				continue
			}

			if (r == ')' || r == '&') && (unicode.IsLetter(nextR) || unicode.IsDigit(nextR)) {
				builder.WriteRune(' ')
				continue
			}

			if r == ',' && !unicode.IsSpace(nextR) && (unicode.IsLetter(nextR) || unicode.IsDigit(nextR)) {
				builder.WriteRune(' ')
				continue
			}
		}
	}

	return builder.String()
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
