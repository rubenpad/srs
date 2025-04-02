export interface IStockRating {
  action: string;
  ticker: string;
  score: number;
  company: string;
  brokerage: string;
  rating_to: string;
  target_to: string;
  target_from: string;
  rating_from: string;
  target_price_change: number;
}

export interface IStockRecommendation {
  time: string;
  score: number;
  rating: string;
  ticker: string;
  buy_ratings: number;
  hold_ratings: number;
  sell_ratings: number;
  target_price_change: number;
  strong_buy_ratings: number;
}

interface IExternalRecommendation {
  buy: number;
  hold: number;
  sell: number;
  symbol: number;
  period: string;
  strongBuy: number;
  strongSell: number;
}

interface IQuote {
  c: number;
  h: number;
  l: number;
  o: number;
  t: number;
  pc: number;
}

export interface IStockDetails {
  keyFacts: string;
  quote: IQuote;
  recommendations: Array<IExternalRecommendation>;
}
