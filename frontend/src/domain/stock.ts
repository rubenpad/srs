export interface IStockRating {
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

export interface IStockRecommendation {
    time: string
    rating: string
    ticker: string,
    buy_ratings: number,
    hold_ratings: number,
    sell_ratings: number,
    target_price_change: number
    strong_buy_ratings: number,
}

interface IExternalRecommendation {
    buy: number,
    hold: number,
    sell: number,
    symbol: number
    period: string,
    strongBuy: number,
    strongSell: number,
}

interface IQuote {
    c: number
    h: number
    l: number
    o: number
    t: number
    pc: number
}

export interface IStockDetails {
    quote: IQuote,
    recommendations: Array<IExternalRecommendation>
}