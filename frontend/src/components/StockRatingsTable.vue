<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { RouterLink } from "vue-router";
import { useStore } from "@/stores/stocks";

enum PageAction {
    PREV,
    NEXT
}

const store = useStore();

const page = ref(1)
const loading = ref(false);
const search = ref('');
const searchTimeout = ref();

const handleSearch = (event: Event) => {
    const value = (event.target as HTMLInputElement).value;
    search.value = value;

    if (searchTimeout.value) {
        clearTimeout(searchTimeout.value);
    }

    searchTimeout.value = setTimeout(async () => {
        loading.value = true
        page.value = 1;
        await store.fetchStockRatings(page.value, search.value);
        loading.value = false;
    }, 300);
};

const handlePageChange = async (pageAction: PageAction) => {
    loading.value = true;
    if (pageAction === PageAction.NEXT) { page.value++ } else { page.value-- }
    await store.fetchStockRatings(page.value, search.value);
    loading.value = false;
}

onMounted(async () => {
    loading.value = true;
    await store.fetchStockRatings();
    loading.value = false;
})

onUnmounted(() => { clearTimeout(searchTimeout.value) })

</script>

<template>
    <div class="flex flex-wrap -mx-3">
        <div class="flex-none w-full max-w-full px-3">
            <div
                class="relative flex flex-col w-full min-w-0 mb-0 break-words bg-white border-0 border-transparent border-solid shadow-soft-xl rounded-2xl bg-clip-border">
                <div class="pt-4 pl-3 mb-5 bg-white rounded-t-md font-bold font-sans uppercase flex flex-row">
                    <h6>Stock Ratings</h6>
                    <div class="flex items-center mt-2 grow sm:mt-0 sm:mr-6 md:mr-0 lg:flex lg:basis-auto">
                        <div class="flex items-center md:ml-auto md:pr-4">
                            <div
                                class="relative flex flex-wrap items-stretch w-full transition-all rounded-lg ease-soft">
                                <span
                                    class="text-sm ease-soft leading-5.6 absolute z-50 -ml-px flex h-full items-center whitespace-nowrap rounded-lg rounded-tr-none rounded-br-none border border-r-0 border-transparent bg-transparent py-2 px-2.5 text-center font-normal text-slate-500 transition-all">
                                </span>
                                <input type="text" v-model="search" @input="handleSearch"
                                    class="pl-4 text-sm focus:shadow-soft-primary-outline ease-soft w-1/100 leading-5.6 relative -ml-px block min-w-0 flex-auto rounded-lg border border-solid border-gray-300 bg-white bg-clip-padding py-2 pr-3 text-gray-700 transition-all placeholder:text-gray-500 focus:border-fuchsia-300 focus:outline-none focus:transition-shadow"
                                    placeholder="Search by ticker..." />
                            </div>
                        </div>
                    </div>
                </div>
                <div class="flex-auto px-0 pt-0 pb-2">
                    <div class="p-0 overflow-x-auto ps">

                        <table class="min-w-full bg-white shadow-md overflow-hidden">
                            <thead class="bg-gray-100">
                                <tr>
                                    <th
                                        class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Brokerage </th>
                                    <th
                                        class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Action </th>
                                    <th
                                        class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Company </th>
                                    <th
                                        class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Rating </th>
                                    <th
                                        class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Price Target</th>
                                    <th
                                        class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        % Price Change</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-gray-200">
                                <tr v-for="stock in store.stockRatings" :key="stock.ticker" class="hover:bg-gray-50">
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ stock.brokerage }}
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ stock.action }}
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                                        <RouterLink :to="`/stock/${stock.ticker}`"
                                            class="text-blue-600 hover:text-blue-800 hover:underline">
                                            {{ `${stock.company} (${stock.ticker})` }}
                                        </RouterLink>
                                    </td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{
                                        `${stock.rating_from} -> ${stock.rating_to}` }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{
                                        `${stock.target_from} -> ${stock.target_to}` }}</td>
                                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{
                                        `${Number((stock.target_price_change * 100).toFixed(2))}%` }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <div class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6">
                        <div class="flex flex-1 justify-between sm:hidden">
                            <button @click="handlePageChange(PageAction.PREV)" :disabled="page === 1 || loading"
                                class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                                :class="{ 'opacity-50 cursor-not-allowed': page === 1 || loading }">
                                Previous
                            </button>
                            <button @click="handlePageChange(PageAction.NEXT)"
                                :disabled="!store.hasMoreStockRatings || loading"
                                class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                                :class="{ 'opacity-50 cursor-not-allowed': !store.hasMoreStockRatings || loading }">
                                Next
                            </button>
                        </div>
                        <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                            <div>
                                <p class="text-sm text-gray-700">
                                    Showing page <span class="font-medium">{{ page }}</span>
                                </p>
                            </div>
                            <div>
                                <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm"
                                    aria-label="Pagination">
                                    <button @click="handlePageChange(PageAction.PREV)" :disabled="page === 1 || loading"
                                        class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0"
                                        :class="{ 'opacity-50 cursor-not-allowed': page === 1 || loading }">
                                        <span class="sr-only">Previous</span>
                                        <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                            <path fill-rule="evenodd"
                                                d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
                                                clip-rule="evenodd" />
                                        </svg>
                                    </button>
                                    <button @click="handlePageChange(PageAction.NEXT)"
                                        :disabled="!store.hasMoreStockRatings || loading"
                                        class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0"
                                        :class="{ 'opacity-50 cursor-not-allowed': !store.hasMoreStockRatings || loading }">
                                        <span class="sr-only">Next</span>
                                        <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                                            <path fill-rule="evenodd"
                                                d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
                                                clip-rule="evenodd" />
                                        </svg>
                                    </button>
                                </nav>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
