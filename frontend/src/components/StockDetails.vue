<script setup lang="ts">
import { onMounted, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useStore } from '@/stores/stocks';
import { Line } from 'vue-chartjs';
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend);

const route = useRoute();
const store = useStore();
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

onMounted(() => store.fetchStockDetails(ticker));
</script>

<template>
    <div class="p-4">
        <div v-if="store.stockDetails.get(ticker)" class="space-y-6">
            <div class="bg-white rounded-lg shadow p-6">
                <h2 class="text-xl font-bold mb-4">{{ `${ticker} Current Quote` }}</h2>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div class="p-4 bg-gray-50 rounded-lg">
                        <p class="text-sm text-gray-600">Current</p>
                        <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.c) }}</p>
                    </div>
                    <div class="p-4 bg-gray-50 rounded-lg">
                        <p class="text-sm text-gray-600">Previous Close</p>
                        <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.pc) }}</p>
                    </div>
                    <div class="p-4 bg-gray-50 rounded-lg">
                        <p class="text-sm text-gray-600">Open</p>
                        <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.o) }}</p>
                    </div>
                    <div class="p-4 bg-gray-50 rounded-lg">
                        <p class="text-sm text-gray-600">High</p>
                        <p class="text-xl font-bold">{{ formatPrice(store.stockDetails.get(ticker)?.quote?.h) }}</p>
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
        <div v-else class="flex items-center justify-center h-64">
            <p class="text-gray-500">Loading stock details...</p>
        </div>
    </div>
</template>