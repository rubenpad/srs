<script setup lang="ts">
import axios from 'axios';
import {computed} from 'vue';
import {Line} from 'vue-chartjs';
import {useRoute} from 'vue-router';
import {useQuery} from '@pinia/colada';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

import type {IStockDetails} from '@/domain/stock';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

const route = useRoute();
const ticker = route.params.ticker as string;

const {isLoading, data} = useQuery<IStockDetails>({
  key: [ticker],
  query: () => axios.get(`/api/stock-details/${ticker}`).then(response => response.data),
});

const formatPrice = (price: number | undefined): string => {
  if (price === undefined) return '$0.00';
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(price);
};

const recommendationData = computed(() => {
  const recommendations = (data.value?.recommendations || []).reverse();
  return {
    labels: recommendations.map(r => r.period),
    datasets: [
      {
        label: 'Buy',
        data: recommendations.map(r => r.buy),
        borderColor: 'rgb(34, 197, 94)',
        tension: 0.1,
      },
      {
        label: 'Hold',
        data: recommendations.map(r => r.hold),
        borderColor: 'rgb(234, 179, 8)',
        tension: 0.1,
      },
      {
        label: 'Sell',
        data: recommendations.map(r => r.sell),
        borderColor: 'rgb(239, 68, 68)',
        tension: 0.1,
      },
    ],
  };
});

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top' as const,
    },
    title: {
      display: true,
      text: `${ticker} - Analyst Recommendations Trends`,
    },
  },
};
</script>

<template>
  <div class="p-4">
    <div v-if="isLoading" class="flex items-center justify-center h-64">
      <p class="text-gray-500">Loading stock details...</p>
    </div>

    <div v-else-if="data" class="space-y-6">
      <RouterLink
        :to="{path: '/', query: $route.query}"
        class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
      >
        <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
          <path
            fill-rule="evenodd"
            d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
            clip-rule="evenodd"
          />
        </svg>
      </RouterLink>
      <div class="bg-white rounded-lg shadow p-6">
        <h2 class="text-xl font-bold mb-4">{{ ticker }}</h2>
        <div class="mb-10">
          <p>{{ data?.keyFacts }}</p>
        </div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="p-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-600">Current</p>
            <p class="text-xl font-bold">{{ formatPrice(data?.quote?.c) }}</p>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-600">Previous Close</p>
            <p class="text-xl font-bold">{{ formatPrice(data?.quote?.pc) }}</p>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-600">Open</p>
            <p class="text-xl font-bold">{{ formatPrice(data?.quote?.o) }}</p>
          </div>
          <div class="p-4 bg-gray-50 rounded-lg">
            <p class="text-sm text-gray-600">High</p>
            <p class="text-xl font-bold">{{ formatPrice(data?.quote?.h) }}</p>
          </div>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow p-6">
        <div class="h-[400px]">
          <Line v-if="data?.recommendations" :data="recommendationData" :options="chartOptions" />
        </div>
      </div>
    </div>

    <div v-else class="flex flex-col items-center justify-center h-64">
      <div class="text-center">
        <h2 class="text-2xl font-bold text-gray-700 mb-2">Stock Details Not Found</h2>
        <p class="text-gray-500 mb-4">We couldn't find any details for ticker "{{ ticker }}"</p>
        <RouterLink
          :to="{path: '/', query: $route.query}"
          class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
        >
          <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path
              fill-rule="evenodd"
              d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
              clip-rule="evenodd"
            />
          </svg>
          Back to Stock Ratings
        </RouterLink>
      </div>
    </div>
  </div>
</template>
