package service

import (
	"context"
	"testing"
	"time"

	"github.com/rubenpad/srs/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStockRatingApi struct {
	mock.Mock
}

type MockStockRatingRepository struct {
	mock.Mock
}

func (m *MockStockRatingRepository) Save(ctx context.Context, stock entity.StockRating) {}

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

func TestUpgradeRatingScore(t *testing.T) {
	ctx := context.Background()
	mockApi := new(MockStockRatingApi)
	mockRepository := new(MockStockRatingRepository)

	testTime := time.Now()
	testStockRatings := []entity.StockRating{
		{
			Brokerage:  "TestBroker",
			Action:     "upgraded by",
			Company:    "TestCompany",
			Ticker:     "TEST",
			RatingFrom: "Hold",
			RatingTo:   "Buy",
			TargetFrom: "$10.00",
			TargetTo:   "$12.00",
			Time:       testTime,
		},
	}

	var savedRatings []entity.StockRating

	mockApi.On("GetStockRatings", ctx, "").
		Return(testStockRatings, "", nil).Once()

	mockRepository.On("BatchSave", ctx, mock.AnythingOfType("[]entity.StockRating")).
		Run(func(args mock.Arguments) {
			savedRatings = args.Get(1).([]entity.StockRating)
		}).Return(nil).Once()

	service := NewStockRatingService(mockRepository, mockApi)

	service.LoadStockRatingsData(ctx)

	mockApi.AssertExpectations(t)
	mockRepository.AssertExpectations(t)

	require.Len(t, savedRatings, 1)
	processedRating := savedRatings[0]

	assert.Equal(t, 5, calculateDateScore(testTime))                                                                                      // Recent date
	assert.Equal(t, 1, calculateTargetPriceChangeScore(calculateTargetPriceChange(processedRating.TargetFrom, processedRating.TargetTo))) // $10 -> $12 is 20% increase
	assert.Equal(t, 5, ratingScaleMap[processedRating.RatingTo])                                                                          // "Buy" rating
	assert.Equal(t, 5, calculateRatingChangeScore(processedRating))                                                                       // Hold -> Buy is an upgrade
	assert.Equal(t, 5, calculateBrokerageActionScore(processedRating))                                                                    // "Upgrade" action

	assert.Equal(t, float32(4.8), processedRating.Score)
}

func TestDowngradeActionScore(t *testing.T) {
	ctx := context.Background()
	mockApi := new(MockStockRatingApi)
	mockRepository := new(MockStockRatingRepository)

	testTime := time.Now()
	testStockRatings := []entity.StockRating{
		{
			Brokerage:  "TestBroker",
			Action:     "downgraded by",
			Company:    "TestCompany",
			Ticker:     "TEST",
			RatingFrom: "Hold",
			RatingTo:   "Sell",
			TargetFrom: "$17.00",
			TargetTo:   "$9.00",
			Time:       testTime,
		},
	}

	var savedRatings []entity.StockRating

	mockApi.On("GetStockRatings", ctx, "").
		Return(testStockRatings, "", nil).Once()

	mockRepository.On("BatchSave", ctx, mock.AnythingOfType("[]entity.StockRating")).
		Run(func(args mock.Arguments) {
			savedRatings = args.Get(1).([]entity.StockRating)
		}).Return(nil).Once()

	service := NewStockRatingService(mockRepository, mockApi)

	service.LoadStockRatingsData(ctx)

	mockApi.AssertExpectations(t)
	mockRepository.AssertExpectations(t)

	require.Len(t, savedRatings, 1)
	processedRating := savedRatings[0]

	assert.Equal(t, 5, calculateDateScore(testTime))
	assert.Equal(t, 0, calculateTargetPriceChangeScore(calculateTargetPriceChange(processedRating.TargetFrom, processedRating.TargetTo)))
	assert.Equal(t, 1, ratingScaleMap[processedRating.RatingTo])
	assert.Equal(t, 1, calculateRatingChangeScore(processedRating))
	assert.Equal(t, 1, calculateBrokerageActionScore(processedRating))

	assert.Equal(t, float32(1.15), processedRating.Score)
}
