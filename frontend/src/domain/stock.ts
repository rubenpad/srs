export interface StockRating {
    brokerage: string
    action: string
    ticker: string
    company: string
    rating_from: string
    rating_to: string,
    target_from: string
    target_to: string
    target_price_change: number
}

export interface StockRecommendation {
    time: string
    "rating": string
    "ticker": string,
    "buy_ratings": number,
    "hold_ratings": number,
    "sell_ratings": number,
    target_price_change: number
    "strong_buy_ratings": number,
}