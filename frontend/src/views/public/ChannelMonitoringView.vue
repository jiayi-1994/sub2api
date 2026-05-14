<template>
  <div class="min-h-screen bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white">
    <header class="border-b border-gray-200 bg-white/95 dark:border-dark-800 dark:bg-dark-900/95">
      <div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-4 sm:px-6">
        <RouterLink to="/home" class="flex min-w-0 items-center gap-3">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-700">
            <img :src="siteLogo || '/logo.png'" alt="Logo" class="h-full w-full object-contain" />
          </span>
          <div class="min-w-0">
            <div class="truncate text-base font-semibold text-gray-950 dark:text-white">
              {{ siteName }}
            </div>
            <div class="truncate text-xs text-gray-500 dark:text-gray-400">
              {{ t('publicMonitoring.subtitle') }}
            </div>
          </div>
        </RouterLink>
        <RouterLink
          to="/login"
          class="inline-flex flex-shrink-0 items-center justify-center rounded-lg bg-primary-600 px-4 py-2 text-sm font-semibold text-white shadow-sm shadow-primary-600/20 transition hover:bg-primary-700"
        >
          {{ t('home.login') }}
        </RouterLink>
      </div>
    </header>

    <main class="mx-auto max-w-6xl px-4 py-6 sm:px-6 lg:py-8">
      <div v-if="disabled" class="rounded-lg border border-gray-200 bg-white p-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400">
        {{ t('publicMonitoring.disabled') }}
      </div>

      <template v-else>
        <MonitorHero
          :overall-status="overallStatus"
          :interval-seconds="DEFAULT_INTERVAL_SECONDS"
          :window="currentWindow"
          :loading="loading"
          :auto-refresh="autoRefresh"
          @update:window="handleWindowChange"
          @refresh="manualReload"
        />

        <MonitorCardGrid
          :items="items as unknown as UserMonitorView[]"
          :window="currentWindow"
          :countdown-seconds="countdown"
          :loading="loading"
          :detail-cache="detailCache as unknown as Record<number, UserMonitorDetail>"
          @card-click="openDetail"
        />

        <MonitorDetailDialog
          :show="showDetail"
          :monitor-id="detailTarget?.id ?? null"
          :title="detailTitle"
          :fetcher="publicDetailFetcher"
          @close="closeDetail"
        />
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { list as listPublic, status as fetchPublicDetail, type PublicMonitorView, type PublicMonitorDetail } from '@/api/publicChannelMonitor'
import type { UserMonitorView, UserMonitorDetail } from '@/api/channelMonitor'
import MonitorHero, {
  type MonitorWindow,
  type OverallStatus,
} from '@/components/user/monitor/MonitorHero.vue'
import MonitorCardGrid from '@/components/user/monitor/MonitorCardGrid.vue'
import MonitorDetailDialog from '@/components/user/MonitorDetailDialog.vue'
import { DEFAULT_INTERVAL_SECONDS, STATUS_OPERATIONAL } from '@/constants/channelMonitor'
import { useAutoRefresh } from '@/composables/useAutoRefresh'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const appStore = useAppStore()

const items = ref<PublicMonitorView[]>([])
const loading = ref(false)
const currentWindow = ref<MonitorWindow>('7d')
const detailCache = reactive<Record<number, PublicMonitorDetail>>({})
const showDetail = ref(false)
const detailTarget = ref<PublicMonitorView | null>(null)

let abortController: AbortController | null = null
let originalDocumentTitle = ''

const disabled = computed(() => appStore.cachedPublicSettings?.channel_monitor_public_enabled === false)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || '', {
  allowRelative: true,
  allowDataUrl: true,
}))

const publicDetailFetcher = async (id: number): Promise<UserMonitorDetail> => {
  const d = await fetchPublicDetail(id)
  return d as unknown as UserMonitorDetail
}

const autoRefresh = useAutoRefresh({
  storageKey: 'public-monitoring-auto-refresh',
  intervals: [30, 60, 120] as const,
  defaultInterval: DEFAULT_INTERVAL_SECONDS,
  onRefresh: () => reload(true),
  shouldPause: () => document.hidden || loading.value,
})
const countdown = autoRefresh.countdown

const overallStatus = computed<OverallStatus>(() => {
  if (items.value.length === 0) return 'operational'
  for (const it of items.value) {
    if (it.primary_status === 'failed' || it.primary_status === 'error') return 'degraded'
    if (it.primary_status !== STATUS_OPERATIONAL) return 'degraded'
  }
  return 'operational'
})

const detailTitle = computed(() => detailTarget.value?.name || t('channelStatus.detailTitle'))

async function reload(silent = false) {
  if (disabled.value) return
  if (abortController) abortController.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  if (!silent) loading.value = true
  try {
    const res = await listPublic({ signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    items.value = res.items || []
  } catch (err: unknown) {
    const e = err as { name?: string; code?: string; response?: { status?: number } }
    if (e?.name === 'AbortError' || e?.code === 'ERR_CANCELED') return
    // 404 = feature disabled server-side after page load; silently surface disabled state.
    if (e?.response?.status === 404) return
    appStore.showError(extractApiErrorMessage(err, t('publicMonitoring.loadError')))
  } finally {
    if (abortController === ctrl) {
      if (!silent) loading.value = false
      countdown.value = DEFAULT_INTERVAL_SECONDS
      abortController = null
    }
  }
}

async function manualReload() {
  await reload(false)
  if (currentWindow.value !== '7d') {
    await Promise.all(items.value.map(it => loadDetail(it.id, true)))
  }
}

async function loadDetail(id: number, force = false) {
  if (!force && detailCache[id]) return
  try {
    detailCache[id] = await fetchPublicDetail(id)
  } catch (err: unknown) {
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.detailLoadError')))
  }
}

async function ensureDetailsForWindow() {
  if (currentWindow.value === '7d') return
  await Promise.all(items.value.map(it => loadDetail(it.id)))
}

async function handleWindowChange(value: MonitorWindow) {
  currentWindow.value = value
  await ensureDetailsForWindow()
}

function openDetail(row: UserMonitorView) {
  detailTarget.value = (items.value.find(it => it.id === row.id) ?? null)
  showDetail.value = true
}

function closeDetail() {
  showDetail.value = false
  detailTarget.value = null
}

watch(items, () => { void ensureDetailsForWindow() })

watch(disabled, (isDisabled) => {
  if (isDisabled) autoRefresh.stop()
  else if (autoRefresh.enabled.value) autoRefresh.start()
})

function applyHeadMeta() {
  if (typeof document === 'undefined') return
  originalDocumentTitle = document.title
  document.title = `${siteName.value} · ${t('publicMonitoring.title')}`
  let meta = document.querySelector('meta[name="robots"]') as HTMLMetaElement | null
  if (!meta) {
    meta = document.createElement('meta')
    meta.setAttribute('name', 'robots')
    document.head.appendChild(meta)
  }
  meta.setAttribute('content', 'noindex, nofollow')
}

function removeHeadMeta() {
  if (typeof document === 'undefined') return
  if (originalDocumentTitle) {
    document.title = originalDocumentTitle
  }
  const meta = document.querySelector('meta[name="robots"]')
  if (meta && meta.getAttribute('content')?.includes('noindex')) {
    meta.parentNode?.removeChild(meta)
  }
}

onMounted(async () => {
  applyHeadMeta()
  await appStore.fetchPublicSettings()
  if (disabled.value) return
  await reload(false)
  autoRefresh.setEnabled(true)
})

onBeforeUnmount(() => {
  removeHeadMeta()
  if (abortController) abortController.abort()
})
</script>
