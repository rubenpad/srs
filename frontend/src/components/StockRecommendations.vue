<script setup lang="ts">
import { onMounted } from "vue";
import { useStore } from "@/stores/stocks";

const store = useStore();
const topRecommendations = 5

onMounted(async () => {
    await store.fetchStockRecommendations(topRecommendations)
});
</script>

<template>
    <div>
        <div class="p-4 pl-3 mb-5 bg-white rounded-t-md font-bold font-sans uppercase">
            <h6>Recommendations</h6>
        </div>
        <div class="flex flex-wrap -mx-3 mb-12">
            <div v-for="stock in store.stockRecommendations" :key="stock.ticker"
                class="w-full max-w-full px-3 mb-6 sm:flex-none xl:mb-0 xl:w-1/5">
                <div
                    class="relative flex flex-col min-w-0 break-words bg-white shadow-soft-xl rounded-md bg-clip-border">
                    <div class="flex-auto">
                        <div class="flex p-4">
                            <RouterLink :to="`/stock/${stock.ticker}`" class="w-full">
                                <h5 class="mb-0 self-center font-bold">
                                    {{ stock.ticker }}
                                    <span :class="['leading-normal text-sm font-weight-bolder', stock.target_price_change >=
                                        0 ? 'text-lime-500' : 'text-red-500'
                                    ]">{{ `${(stock.target_price_change > 0 ? '+' : '') + stock.target_price_change}%`
                                        }}</span>
                                </h5>
                                <span :class="[
                                    'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset mt-2',
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
                            </RouterLink>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
