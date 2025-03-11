<script setup lang="ts">
import StockRatingsTable from "@/components/StockRatingsTable.vue";
import StockRecommendations from '@/components/StockRecommendations.vue'
</script>

<template>
  <div>
    <header class="bg-white shadow-sm w-full">
      <div class="w-full px-6 py-4">
        <h1 class="text-3xl font-bold text-gray-900 font-space-grotesk tracking-tight">
          Stock Rating System
        </h1>
      </div>
    </header>

    <main class="w-full px-6 py-6">
      <StockRecommendations />
      <StockRatingsTable />
    </main>
  </div>
</template>
