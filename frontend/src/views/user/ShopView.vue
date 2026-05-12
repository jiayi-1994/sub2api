<template>
  <AppLayout>
    <div class="shop-page-layout">
      <div class="card flex-1 min-h-0 overflow-hidden">
        <div class="shop-embed-shell">
          <a
            href="https://pay.ldxp.cn/shop/TLLMN09N"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm shop-open-fab"
          >
            <svg class="mr-1.5 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
            </svg>
            新窗口打开
          </a>
          <iframe
            src="https://pay.ldxp.cn/shop/TLLMN09N"
            class="shop-embed-frame"
            allowfullscreen
          ></iframe>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue'
</script>

<style scoped>
.shop-page-layout {
  @apply flex flex-col;
  height: calc(100vh - 64px - 4rem);
}

.shop-embed-shell {
  @apply relative;
  @apply h-full w-full overflow-hidden rounded-2xl;
  @apply bg-gradient-to-b from-gray-50 to-white dark:from-dark-900 dark:to-dark-950;
  @apply p-0;
}

.shop-open-fab {
  @apply absolute right-3 top-3 z-10;
  @apply shadow-sm backdrop-blur supports-[backdrop-filter]:bg-white/80 dark:supports-[backdrop-filter]:bg-dark-800/80;
}

.shop-embed-frame {
  display: block;
  margin: 0;
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 0;
  box-shadow: none;
  background: transparent;
}
</style>
