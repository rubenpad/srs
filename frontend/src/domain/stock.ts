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