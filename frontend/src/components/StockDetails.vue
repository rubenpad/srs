<script setup lang="ts">
import { onMounted, computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useStore } from '@/stores/stocks';
import { Line } from 'vue-chartjs';
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

const route = useRoute();
const store = useStore();
const loading = ref(false);
const ticker = route.params.ticker as string;

const formatPrice = (price: number | undefined): string => {
    if (price === undefined) return '$0.00';
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(price);
};

const recommendationData = computed(() => {
    const recommendations = (store.stockDetails.get(ticker)?.recommendations || []).reverse();
    return {
        labels: recommendations.map(r => r.period),
        datasets: [
            {
                label: 'Buy',
                data: recommendations.map(r => r.buy),
                borderColor: 'rgb(34, 197, 94)',
                tension: 0.1
            },
            {
                label: 'Hold',
                data: recommendations.map(r => r.hold),
                borderColor: 'rgb(234, 179, 8)',
                tension: 0.1
            },
            {
                label: 'Sell',
                data: recommendations.map(r => r.sell),
                borderColor: 'rgb(239, 68, 68)',
                tension: 0.1
            }
        ]
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
            text: `${ticker} - Analyst Recommendations Trends`
        }
    }
};

onMounted(async () => {
    loading.value = true;
    await store.fetchStockDetails(ticker)
    loading.value = false;
});
</script>

<template>
    <div class="p-4">
        <div v-if="loading" class="flex items-center justify-center h-64">
            <p class="text-gray-500">Loading stock details...</p>
        </div>

        <div v-else-if="store.stockDetails.get(ticker)" class="space-y-6">
            <RouterLink to="/"
                class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                < </RouterLink>
                    <div class="bg-white rounded-lg shadow p-6">
                        <h2 class="text-xl font-bold mb-4">{{ `${ticker} Current Quote` }}</h2>
                        <div class="mb-10">
                            <p>{{  store.stockDetails.get(ticker)?.keyFacts }}</p>
                        </div>
                        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                            <div class="p-4 bg-gray-50 rounded-lg">
                                <p class="text-sm text-gray-600">Current</p>
                                <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.c) }}
                                </p>
                            </div>
                            <div class="p-4 bg-gray-50 rounded-lg">
                                <p class="text-sm text-gray-600">Previous Close</p>
                                <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.pc)
                                }}</p>
                            </div>
                            <div class="p-4 bg-gray-50 rounded-lg">
                                <p class="text-sm text-gray-600">Open</p>
                                <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.o) }}
                                </p>
                            </div>
                            <div class="p-4 bg-gray-50 rounded-lg">
                                <p class="text-sm text-gray-600">High</p>
                                <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.h) }}
                                </p>
                            </div>
                        </div>
                    </div>

                    <div class="bg-white rounded-lg shadow p-6">
                        <div class="h-[400px]">
                            <Line v-if="store.stockDetails.get(ticker)?.recommendations" :data="recommendationData"
                                :options="chartOptions" />
                        </div>
                    </div>
        </div>

        <div v-else class="flex flex-col items-center justify-center h-64">
            <div class="text-center">
                <h2 class="text-2xl font-bold text-gray-700 mb-2">Stock Details Not Found</h2>
                <p class="text-gray-500 mb-4">
                    We couldn't find any details for ticker "{{ ticker }}"
                </p>
                <RouterLink to="/"
                    class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
                    ← Back to Stock Ratings
                </RouterLink>
            </div>
        </div>
    </div>
</template>