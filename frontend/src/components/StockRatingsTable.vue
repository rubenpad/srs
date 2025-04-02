<script setup lang="ts">
import axios from 'axios';
import {useQuery} from '@pinia/colada';
import {ref, onUnmounted, watch} from 'vue';
import {RouterLink, useRoute, useRouter} from 'vue-router';

import type {IStockRating} from '@/domain/stock';

enum PageAction {
  PREV,
  NEXT,
}

const route = useRoute();
const router = useRouter();

const page = ref(parseInt((route.query.page as string) || '1'));
const search = ref((route.query.search as string) || '');
const pages = ref(new Map());
const searchTimeout = ref();
const stockRatingsPageSize = 10;
const hasMoreStockRatings = ref(true);

const fetchStockRatings = async () => {
  const nextPageValue = pages.value.get(page.value - 1) || '';
  const response = await axios.get(`/api/stock-ratings`, {
    params: {
      search: search.value,
      nextPage: nextPageValue,
      pageSize: stockRatingsPageSize,
    },
  });

  hasMoreStockRatings.value = response.data.nextPage !== '';
  if (hasMoreStockRatings.value) pages.value.set(page.value, response.data.nextPage);

  return response.data;
};

const {isLoading, data, status, refetch} = useQuery<{
  data: Array<IStockRating>;
  nextPage: string;
}>({
  key: () => ['stock-ratings', page.value],
  query: fetchStockRatings,
  placeholderData: placeholderData => placeholderData,
});

const handleSearch = (event: Event) => {
  const value = (event.target as HTMLInputElement).value;
  search.value = value;

  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value);
  }

  searchTimeout.value = setTimeout(async () => {
    page.value = 1;
    router.replace({
      query: {...route.query, page: 1, search: search.value},
    });
    await refetch();
  }, 300);
};

const handleRefetch = () => {
  refetch();
};

const handlePageChange = async (pageAction: PageAction) => {
  if (pageAction === PageAction.NEXT) {
    page.value++;
  } else {
    page.value--;
  }
  router.replace({
    query: {...route.query, page: page.value, search: search.value},
  });
  await refetch();
};

onUnmounted(() => {
  clearTimeout(searchTimeout.value);
});

watch(
  () => route.query,
  query => {
    if (query.search && query.search !== search.value) {
      search.value = query.search as string;
    }

    if (query.page && parseInt(query.page as string) !== page.value) {
      page.value = parseInt(query.page as string);
    }
  },
);
</script>

<template>
  <div class="flex flex-wrap -mx-3">
    <div class="flex-none w-full max-w-full px-3">
      <div
        class="relative flex flex-col w-full min-w-0 mb-0 break-words bg-white border-0 border-transparent border-solid shadow-soft-xl rounded-2xl bg-clip-border"
      >
        <div class="pt-4 pl-3 mb-5 bg-white rounded-t-md font-bold font-sans uppercase flex flex-row">
          <h6>Stock Ratings</h6>
          <div class="flex items-center mt-2 grow sm:mt-0 sm:mr-6 md:mr-0 lg:flex lg:basis-auto">
            <div class="flex items-center md:ml-auto md:pr-4">
              <div class="relative flex flex-wrap items-stretch w-full transition-all rounded-lg ease-soft">
                <span
                  class="text-sm ease-soft leading-5.6 absolute z-50 -ml-px flex h-full items-center whitespace-nowrap rounded-lg rounded-tr-none rounded-br-none border border-r-0 border-transparent bg-transparent py-2 px-2.5 text-center font-normal text-slate-500 transition-all"
                >
                </span>
                <div class="relative flex flex-wrap items-stretch w-full transition-all rounded-lg ease-soft">
                  <span
                    class="text-sm ease-soft leading-5.6 absolute z-50 -ml-px flex h-full items-center whitespace-nowrap rounded-lg rounded-tr-none rounded-br-none border border-r-0 border-transparent bg-transparent py-2 px-2.5 text-center font-normal text-slate-500 transition-all"
                  >
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      class="w-5 h-5 text-gray-500"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"
                      />
                    </svg>
                  </span>
                  <input
                    type="text"
                    v-model="search"
                    @input="handleSearch"
                    class="pl-12 text-sm focus:shadow-soft-primary-outline ease-soft w-1/100 leading-5.6 relative -ml-px block min-w-0 flex-auto rounded-lg border border-solid border-gray-300 bg-white bg-clip-padding py-2 pr-3 text-gray-700 transition-all placeholder:text-gray-500 focus:border-fuchsia-300 focus:outline-none focus:transition-shadow"
                    placeholder="Search by ticker..."
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="flex-auto px-0 pt-0 pb-2">
          <div class="p-0 overflow-x-auto ps">
            <table class="min-w-full bg-white shadow-md overflow-hidden">
              <thead class="bg-gray-100">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Brokerage
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Company
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rating</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Price Target
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    % Price Change
                  </th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Score</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200">
                <template v-if="isLoading">
                  <tr v-for="row in 10" :key="row" class="animate-pulse">
                    <td v-for="column in 7" :key="column" class="px-6 py-4 whitespace-nowrap">
                      <div class="h-4 bg-gray-200 rounded w-3/4"></div>
                    </td>
                  </tr>
                </template>

                <template v-else-if="status === 'error'">
                  <tr>
                    <td colspan="7" class="text-center py-8">
                      <div class="bg-white rounded-md shadow-soft-xl p-6 max-w-md mx-auto">
                        <h2 class="text-xl font-bold text-red-600 mb-2">Failed to load stock ratings</h2>
                        <p class="text-gray-600 mb-4">
                          We couldn't load the stock ratings information. Please try again.
                        </p>
                        <button
                          @click="handleRefetch"
                          class="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                        >
                          Try Again
                        </button>
                      </div>
                    </td>
                  </tr>
                </template>

                <tr v-else-if="status === 'success' && data?.data?.length === 0">
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500" colspan="7">No stock ratings found</td>
                </tr>

                <tr v-else v-for="stock in data?.data" :key="stock.ticker" class="hover:bg-gray-50">
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ stock.brokerage }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ stock.action }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    <RouterLink
                      :to="{
                        path: `/stock/${stock.ticker}`,
                        query: $route.query,
                      }"
                      class="text-blue-600 hover:text-blue-800 hover:underline"
                    >
                      {{ `${stock.company} (${stock.ticker})` }}
                    </RouterLink>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ `${stock.rating_from} -> ${stock.rating_to}` }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ `${stock.target_from} -> ${stock.target_to}` }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ `${Number((stock.target_price_change * 100).toFixed(2))}%` }}
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{ stock.score }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6">
            <div class="flex flex-1 justify-between sm:hidden">
              <button
                @click="handlePageChange(PageAction.PREV)"
                :disabled="page === 1 || isLoading"
                class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                :class="{
                  'opacity-50 cursor-not-allowed': page === 1 || isLoading,
                }"
              >
                Previous
              </button>
              <button
                @click="handlePageChange(PageAction.NEXT)"
                :disabled="!hasMoreStockRatings || isLoading"
                class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                :class="{
                  'opacity-50 cursor-not-allowed': !hasMoreStockRatings || isLoading,
                }"
              >
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
                <nav class="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
                  <button
                    @click="handlePageChange(PageAction.PREV)"
                    :disabled="page === 1 || isLoading"
                    class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0"
                    :class="{
                      'opacity-50 cursor-not-allowed': page === 1 || isLoading,
                    }"
                  >
                    <span class="sr-only">Previous</span>
                    <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                      <path
                        fill-rule="evenodd"
                        d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
                        clip-rule="evenodd"
                      />
                    </svg>
                  </button>
                  <button
                    @click="handlePageChange(PageAction.NEXT)"
                    :disabled="!hasMoreStockRatings || isLoading"
                    class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20 focus:outline-offset-0"
                    :class="{
                      'opacity-50 cursor-not-allowed': !hasMoreStockRatings || isLoading,
                    }"
                  >
                    <span class="sr-only">Next</span>
                    <svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                      <path
                        fill-rule="evenodd"
                        d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
                        clip-rule="evenodd"
                      />
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
    opacity: 0.5;
  }
}
</style>
