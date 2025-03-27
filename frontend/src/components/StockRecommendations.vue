<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useStore } from "@/stores/stocks";

const store = useStore();
const loading = ref(false);
const topRecommendations = 5

onMounted(async () => {
    loading.value = true;
    await store.fetchStockRecommendations(topRecommendations)
    loading.value = false;
});
</script>

<template>
    <div>
        <div class="p-4 pl-3 mb-5 bg-white rounded-t-md font-bold font-sans uppercase">
            <h6>Recommendations</h6>
        </div>
        <div class="flex flex-wrap -mx-3 mb-12 gap-y-4">
            <template v-if="loading">
                <div v-for="placeholder in topRecommendations" :key="placeholder"
                    class="w-full max-w-full px-3 mb-6 sm:flex-none xl:mb-0 xl:w-1/5">
                    <div
                        class="relative flex flex-col min-w-0 break-words bg-white shadow-soft-xl rounded-md bg-clip-border">
                        <div class="flex-auto animate-pulse">
                            <div class="grid grid-flow-col grid-rows-2 p-4">
                                <!-- Ticker and Price Change Placeholder -->
                                <div class="mb-0 self-center">
                                    <div class="h-6 bg-gray-200 rounded w-20 mb-2"></div>
                                    <div class="h-4 bg-gray-200 rounded w-12"></div>
                                </div>
                                <!-- Rating Placeholder -->
                                <div class="mt-2">
                                    <div class="h-10 bg-gray-200 rounded w-full"></div>
                                </div>
                                <!-- Score Placeholder -->
                                <div class="row-start-1 row-end-3 text-center p-8">
                                    <div class="h-5 bg-gray-200 rounded w-16 mx-auto mb-2"></div>
                                    <div class="h-6 bg-gray-200 rounded w-10 mx-auto"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>

            <div v-else v-for="stock in store.stockRecommendations" :key="stock.ticker"
                class="w-full max-w-full px-3 mb-6 sm:flex-none xl:mb-0 xl:w-1/5">
                <div
                    class="relative flex flex-col min-w-0 break-words bg-white shadow-soft-xl rounded-md bg-clip-border">
                    <div class="flex-auto">
                        <RouterLink :to="`/stock/${stock.ticker}`" class="w-full">
                            <div class="grid grid-flow-col grid-rows-2 p-4">
                                <h5 class="mb-0 self-center font-bold">
                                    {{ stock.ticker }}
                                    <span :class="['leading-normal text-sm font-weight-bolder', stock.target_price_change >=
                                        0 ? 'text-lime-500' : 'text-red-500'
                                    ]">{{ `${(stock.target_price_change > 0 ? '+' : '') + stock.target_price_change}%`
                                    }}</span>
                                </h5>
                                <span :class="[
                                    'rounded-md text-md font-medium ring-1 ring-inset mt-2 text-center p-2',
                                    {
                                        'bg-green-50 text-green-700 ring-green-600/20': stock.rating === 'Strong Buy',
                                        'bg-lime-50 text-lime-700 ring-lime-600/20': stock.rating === 'Buy',
                                        'bg-yellow-50 text-yellow-700 ring-yellow-600/20': stock.rating === 'Hold',
                                        'bg-orange-50 text-orange-700 ring-orange-600/20': stock.rating === 'Sell',
                                        'bg-red-50 text-red-700 ring-red-600/20': stock.rating === 'Strong Sell'
                                    }
                                ]">
                                    {{ stock.rating }}
                                </span>
                                <div class="row-start-1 row-end-3 text-center p-8">
                                    <h5>
                                        Score
                                    </h5>
                                    <span class="text-lg font-bold">{{
                                        stock.score }}</span>
                                </div>

                            </div>
                        </RouterLink>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse {

    0%,
    100% {
        opacity: 1;
    }

    50% {
        opacity: .5;
    }
}
</style>
