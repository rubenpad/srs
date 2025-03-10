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
    <div class="flex flex-wrap -mx-3 my-12">
        <div v-for="stock in store.stockRecommendations" :key="stock.ticker"
            class="w-full max-w-full px-3 mb-6 sm:w-1/2 sm:flex-none xl:mb-0 xl:w-1/5">
            <div class="relative flex flex-col min-w-0 break-words bg-white shadow-soft-xl rounded-md bg-clip-border">
                <div class="flex-auto">
                    <div class="flex p-4">
                        <RouterLink :to="`/stock/${stock.ticker}`">
                            <h5 class="mb-0 self-center font-bold">
                                {{ stock.ticker }}
                                <span :class="['leading-normal text-sm font-weight-bolder', stock.target_price_change >=
                                    0 ? 'text-lime-500' : 'text-red-500'
                                ]">{{ `${(stock.target_price_change > 0 ? '+' : '') + stock.target_price_change}%`
                                    }}</span>
                            </h5>
                            <span
                                class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-green-600/20 ring-inset">{{
                                    stock.rating }}</span>
                        </RouterLink>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
