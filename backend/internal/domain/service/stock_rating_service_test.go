package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rubenpad/srs/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStockRatingApi struct {
	mock.Mock
}

type MockStockRatingRepository struct {
	mock.Mock
}

func (m *MockStockRatingRepository) Save(ctx context.Context, stock entity.StockRating) {
	m.Called(ctx, stock)
}

func (m *MockStockRatingRepository) BatchSave(ctx context.Context, stockRatings []entity.StockRating) {
	m.Called(ctx, stockRatings)
}

func (m *MockStockRatingRepository) GetStockRatings(ctx context.Context, nextPage string, pageSize int, search string) ([]entity.StockRating, error) {
	args := m.Called(ctx, nextPage, pageSize, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.StockRating), args.Error(1)
}

func (m *MockStockRatingRepository) GetStockRecommendations(ctx context.Context, pageSize int) ([]entity.StockRatingAggregate, error) {
	args := m.Called(ctx, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.StockRatingAggregate), args.Error(1)
}

func (m *MockStockRatingApi) GetStockDetails(ctx context.Context, ticker string) *entity.StockDetails {
	args := m.Called(ctx, ticker)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*entity.StockDetails)
}

func (m *MockStockRatingApi) GetStockRatings(ctx context.Context, nextPage string) ([]entity.StockRating, string, error) {
	args := m.Called(ctx, nextPage)
	return args.Get(0).([]entity.StockRating), args.String(1), args.Error(2)
}

func TestLoadStockRatingsData(t *testing.T) {
	ctx := context.Background()
	mockApi := new(MockStockRatingApi)
	mockRepository := new(MockStockRatingRepository)

	testTime := time.Now()

	testBatch1 := []entity.StockRating{
		{
			Brokerage:  "TestBroker1",
			Action:     "upgraded by",
			Company:    "TestCompany1",
			Ticker:     "TEST1",
			RatingFrom: "Hold",
			RatingTo:   "Buy",
			TargetFrom: "$10.00",
			TargetTo:   "$15.00",
			Time:       testTime,
		},
	}

	testBatch2 := []entity.StockRating{
		{
			Brokerage:  "TestBroker2",
			Action:     "downgraded by",
			Company:    "TestCompany2",
			Ticker:     "TEST2",
			RatingFrom: "Buy",
			RatingTo:   "Sell",
			TargetFrom: "$20.00",
			TargetTo:   "$15.00",
			Time:       testTime,
		},
	}

	var mu sync.Mutex
	var processedRatings []entity.StockRating

	mockApi.On("GetStockRatings", ctx, "").
		Return(testBatch1, "next_page", nil).Once()
	mockApi.On("GetStockRatings", ctx, "next_page").
		Return(testBatch2, "", nil).Once()

	// Capture processed ratings in thread-safe way
	mockRepository.On("Save", ctx, mock.AnythingOfType("entity.StockRating")).
		Run(func(args mock.Arguments) {
			mu.Lock()
			processedRatings = append(processedRatings, args.Get(1).(entity.StockRating))
			mu.Unlock()
		}).Return(nil)

	service := NewStockRatingService(mockRepository, mockApi)

	service.LoadStockRatingsData(ctx)

	mockApi.AssertExpectations(t)
	mockRepository.AssertExpectations(t)

	assert.Equal(t, len(testBatch1)+len(testBatch2), len(processedRatings))

	var upgradedRating entity.StockRating
	for _, r := range processedRatings {
		if r.Ticker == "TEST1" {
			upgradedRating = r
			break
		}
	}

	assert.NotEmpty(t, upgradedRating)
	assert.Equal(t, 5, calculateDateScore(upgradedRating.Time))
	assert.Equal(t, 5, calculateTargetPriceChangeScore(calculateTargetPriceChange(upgradedRating.TargetFrom, upgradedRating.TargetTo)))
	assert.Equal(t, 5, ratingScaleMap[upgradedRating.RatingTo])
	assert.Equal(t, 5, calculateRatingChangeScore(upgradedRating))
	assert.Equal(t, 5, calculateBrokerageActionScore(upgradedRating))

	var downgradedRating entity.StockRating
	for _, r := range processedRatings {
		if r.Ticker == "TEST2" {
			downgradedRating = r
			break
		}
	}

	assert.NotEmpty(t, downgradedRating)
	assert.Equal(t, 5, calculateDateScore(downgradedRating.Time))
	assert.Equal(t, 0, calculateTargetPriceChangeScore(calculateTargetPriceChange(downgradedRating.TargetFrom, downgradedRating.TargetTo)))
	assert.Equal(t, 1, ratingScaleMap[downgradedRating.RatingTo])
	assert.Equal(t, 1, calculateRatingChangeScore(downgradedRating))
	assert.Equal(t, 1, calculateBrokerageActionScore(downgradedRating))
}
