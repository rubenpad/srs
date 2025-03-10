import axios from "axios";
import { defineStore } from 'pinia'

import type { StockRating } from "@/domain/stock";

export const useStore = defineStore("stocks", {
  state: () => ({
    stockRatingsPages: new Map(),
    hasMoreStockRatings: true,
    stockRatingsPageSize: 10,
    stockRatings: [] as Array<StockRating>,
    stockDetails: {} as Record<string, object>,
    stockRecommendations: [],
  }),

  actions: {
    async fetchStockRatings(page = 1, search = '') {
      const nextPageValue = this.stockRatingsPages.get(page - 1) || ''

      const response = await axios
        .get(`/api/stock-ratings`,
          { params: { nextPage: nextPageValue, pageSize: this.stockRatingsPageSize, search } });

      this.stockRatings = response.data.data;
      this.hasMoreStockRatings = response.data.nextPage !== ''

      if (this.hasMoreStockRatings) this.stockRatingsPages.set(page, response.data.nextPage)
    },

    async fetchStockRecommendations(pageSize: number) {
      const response = await axios.get(`/api/stock-recommendations?pageSize=${pageSize}`);
      this.stockRecommendations = response.data.data;
    },

    async fetchStockDetails(ticker: string) {
      const { data } = await axios.get(`https://query1.finance.yahoo.com/v7/finance/quote?symbols=${ticker}`);
      this.stockDetails[ticker] = data.quoteResponse.result[0];
    },
  },
});
