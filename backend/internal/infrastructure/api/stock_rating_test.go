package api

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rubenpad/srs/internal/domain/entity"
)

var testCases = []struct {
	expectError bool
	name        string
	line        string
	errorMsg    string
	expected    entity.StockRating
}{
	{
		name: "Standard Case - Reiterated Buy",
		line: "MOMO$13.00$13.00HelloGroupreiteratedbyBenchmarkBuyBuyFriMar14202500:30UTC",
		expected: entity.StockRating{
			Ticker:     "MOMO",
			TargetFrom: "$13.00",
			TargetTo:   "$13.00",
			Company:    "Hello Group",
			Action:     "reiterated by",
			Brokerage:  "Benchmark",
			RatingFrom: "Buy",
			RatingTo:   "Buy",
			Time:       mustParseTime("FriMar14202500:30UTC"),
		},
		expectError: false,
	},
	{
		name: "Target Lowered - Market Perform",
		line: "BLND$3.85$3.50BlendLabstargetloweredbyKeefe,Bruyette&WoodsMarketPerformMarketPerformTueMar04202500:30UTC",
		expected: entity.StockRating{
			Ticker:     "BLND",
			TargetFrom: "$3.85",
			TargetTo:   "$3.50",
			Company:    "Blend Labs",
			Action:     "target lowered by",
			Brokerage:  "Keefe, Bruyette & Woods",
			RatingFrom: "Market Perform",
			RatingTo:   "Market Perform",
			Time:       mustParseTime("TueMar04202500:30UTC"),
		},
		expectError: false,
	},
	{
		name: "Upgraded - Sell to Neutral",
		line: "BSBR$4.20$4.70BancoSantander(Brasil)upgradedbyTheGoldmanSachsGroupSellNeutralThuFeb13202500:30UTC",
		expected: entity.StockRating{
			Ticker:     "BSBR",
			TargetFrom: "$4.20",
			TargetTo:   "$4.70",
			Company:    "Banco Santander (Brasil)",
			Action:     "upgraded by",
			Brokerage:  "The Goldman Sachs Group",
			RatingFrom: "Sell",
			RatingTo:   "Neutral",
			Time:       mustParseTime("ThuFeb13202500:30UTC"),
		},
		expectError: false,
	},
	{
		name: "Initiated Buy - Multi-word Brokerage",
		line: "AKBA$6.00$6.00AkebiaTherapeuticsinitiatedbyJefferiesFinancialGroupBuyBuyWedApr02202500:30UTC",
		expected: entity.StockRating{
			Ticker:     "AKBA",
			TargetFrom: "$6.00",
			TargetTo:   "$6.00",
			Company:    "Akebia Therapeutics",
			Action:     "initiated by",
			Brokerage:  "Jefferies Financial Group",
			RatingFrom: "Buy",
			RatingTo:   "Buy",
			Time:       mustParseTime("WedApr02202500:30UTC"),
		},
		expectError: false,
	},
	{
		name: "Initiated Buy - Multi-word Brokerage",
		line: "AKBA$1,500.00$1,500.00AkebiaTherapeuticsinitiatedbyJefferiesFinancialGroupBuyBuyWedApr02202500:30UTC",
		expected: entity.StockRating{
			Ticker:     "AKBA",
			TargetFrom: "$1,500.00",
			TargetTo:   "$1,500.00",
			Company:    "Akebia Therapeutics",
			Action:     "initiated by",
			Brokerage:  "Jefferies Financial Group",
			RatingFrom: "Buy",
			RatingTo:   "Buy",
			Time:       mustParseTime("WedApr02202500:30UTC"),
		},
		expectError: false,
	},
	{
		name:        "Error Case - Invalid Date",
		line:        "MOMO$13.00$13.00HelloGroupreiteratedbyBenchmarkBuyBuyInvalidDate",
		expectError: true,
		errorMsg:    "invalid stock rating format: date not found",
	},
	{
		name:        "Error Case - Missing Action",
		line:        "MOMO$13.00$13.00HelloGroupBenchmarkBuyBuyFriMar14202500:30UTC",
		expectError: true,
		errorMsg:    "action not found",
	},
	{
		name:        "Error Case - Missing Ticker/Targets",
		line:        "InvalidStartHelloGroupreiteratedbyBenchmarkBuyBuyFriMar14202500:30UTC",
		expectError: true,
		errorMsg:    "invalid stock rating format: ticker/targets not matched",
	},
	{
		name:        "Error Case - Missing Rating From",
		line:        "MOMO$13.00$13.00HelloGroupreiteratedbyBenchmarkMarketPerformFriMar14202500:30UTC",
		expectError: true,
		errorMsg:    "ratingFrom not found",
	},
	{
		name:        "Error Case - Malformed Ticker Part",
		line:        "123$13.00$13.00HelloGroupreiteratedbyBenchmarkBuyBuyFriMar14202500:30UTC",
		expectError: true,
		errorMsg:    "ticker/targets not matched",
	},
}

func TestParseStockRatingLine(t *testing.T) {

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseStockRatingLine(tc.line)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				} else if tc.errorMsg != "" && !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error message containing '%s', but got '%v'", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}

				if !reflect.DeepEqual(actual, tc.expected) {
					t.Errorf("Mismatch:\nExpected: %+v\nActual:   %+v", tc.expected, actual)
				}
			}
		})
	}
}

func mustParseTime(value string) time.Time {
	t, err := time.Parse("MonJan02200615:04MST", value)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse time '%s': %v", value, err))
	}
	return t.Truncate(24 * time.Hour)
}
