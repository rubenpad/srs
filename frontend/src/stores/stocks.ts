import axios from "axios";
import { defineStore } from 'pinia'

import type { IStockRating, IStockRecommendation, IStockDetails } from "@/domain/stock";

export const useStore = defineStore("stocks", {
  state: () => ({
    stockDetails: new Map<string, IStockDetails>(),
    stockRatingsPageSize: 10,
    hasMoreStockRatings: true,
    stockRatingsPages: new Map(),
    stockRatings: [] as Array<IStockRating>,
    stockRecommendations: [] as Array<IStockRecommendation>,
  }),

  actions: {
    async fetchStockRatings(page = 1, search = '') {
      try {
        const nextPageValue = this.stockRatingsPages.get(page - 1) || ''

        const response = await axios
          .get(`/api/stock-ratings`,
            { params: { nextPage: nextPageValue, pageSize: this.stockRatingsPageSize, search } });

        this.stockRatings = response.data.data;
        this.hasMoreStockRatings = response.data.nextPage !== ''

        if (this.hasMoreStockRatings) this.stockRatingsPages.set(page, response.data.nextPage)
      } catch {
        return null;
      }

    },

    async fetchStockRecommendations(pageSize: number) {
      try {
        const response = await axios.get(`/api/stock-recommendations?pageSize=${pageSize}`);
        this.stockRecommendations = response.data.data;
      } catch  {
        return null;
      }
    },

    async fetchStockDetails(ticker: string) {
      const maybeDetails = this.stockDetails.get(ticker)
      if (maybeDetails !== undefined) return;

      try {
        const response = await axios.get(`/api/stock-details/${ticker}`);
        this.stockDetails.set(ticker, response.data)
      } catch  {
        return null;
      }
    },
  },
});
