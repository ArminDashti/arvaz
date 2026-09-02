<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { api, type HostMetrics } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const selectClass = 'rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground'

const metrics = ref<HostMetrics | null>(null)
const error = ref('')
const loading = ref(false)
const refreshSeconds = ref(2)
const refreshChoices = [1, 2, 4] as const
let refreshTimer: ReturnType<typeof setInterval> | undefined

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const res = await api.hostMetrics()
    if (res.error) {
      error.value = res.error
    } else {
      error.value = ''
      metrics.value = res
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

function armRefresh() {
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = setInterval(() => {
    void load()
  }, refreshSeconds.value * 1000)
}

onMounted(() => {
  void load()
  armRefresh()
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})

watch(refreshSeconds, armRefresh)

function formatGb(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n.toFixed(2)} GB`
}

function formatMbps(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n >= 10 ? n.toFixed(1) : n.toFixed(2)} Mbps`
}

function formatPct(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n.toFixed(1)}%`
}

function barWidth(pct: number | undefined): string {
  const v = Math.max(0, Math.min(100, pct ?? 0))
  return `${v}%`
}

function usagePct(used: number | undefined, total: number | undefined): number {
  if (!total || total <= 0 || used == null) return 0
  return (used / total) * 100
}

const memoryPct = computed(() =>
  usagePct(metrics.value?.memory?.usedGb, metrics.value?.memory?.totalGb),
)
const diskPct = computed(() => usagePct(metrics.value?.disk?.usedGb, metrics.value?.disk?.totalGb))
</script>

<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card>
      <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
        <div>
          <CardTitle>Host</CardTitle>
          <p class="mt-1 text-sm text-muted-foreground">
            Docker host CPU, memory, disk, and network. Updates only while this page is open.
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <label class="flex items-center gap-2 text-sm text-muted-foreground">
            Interval
            <select v-model.number="refreshSeconds" :class="selectClass">
              <option v-for="n in refreshChoices" :key="n" :value="n">{{ n }}s</option>
            </select>
          </label>
          <button type="button" class="field-button" :disabled="loading" @click="load">
            {{ loading ? 'Refreshing…' : 'Refresh' }}
          </button>
        </div>
      </CardHeader>
      <CardContent class="space-y-6">
        <section>
          <h2 class="mb-3 text-sm font-medium text-foreground">CPU (per core)</h2>
          <div v-if="!metrics?.cpuCores?.length" class="text-sm text-muted-foreground">—</div>
          <ul v-else class="space-y-2">
            <li
              v-for="(pct, i) in metrics.cpuCores"
              :key="i"
              class="grid grid-cols-[4rem_1fr_4rem] items-center gap-2 text-sm"
            >
              <span class="text-muted-foreground">Core {{ i }}</span>
              <div class="h-2 overflow-hidden rounded bg-muted">
                <div class="h-full bg-primary transition-[width] duration-300" :style="{ width: barWidth(pct) }" />
              </div>
              <span class="text-right tabular-nums">{{ formatPct(pct) }}</span>
            </li>
          </ul>
        </section>

        <section class="grid gap-4 sm:grid-cols-2">
          <div>
            <h2 class="mb-2 text-sm font-medium text-foreground">Memory</h2>
            <div class="mb-1 h-2 overflow-hidden rounded bg-muted">
              <div class="h-full bg-primary transition-[width] duration-300" :style="{ width: barWidth(memoryPct) }" />
            </div>
            <p class="text-sm tabular-nums text-muted-foreground">
              {{ formatGb(metrics?.memory?.usedGb) }} / {{ formatGb(metrics?.memory?.totalGb) }}
              <span class="ml-2">(avail {{ formatGb(metrics?.memory?.availableGb) }})</span>
            </p>
          </div>
          <div>
            <h2 class="mb-2 text-sm font-medium text-foreground">Disk</h2>
            <div class="mb-1 h-2 overflow-hidden rounded bg-muted">
              <div class="h-full bg-primary transition-[width] duration-300" :style="{ width: barWidth(diskPct) }" />
            </div>
            <p class="text-sm tabular-nums text-muted-foreground">
              {{ formatGb(metrics?.disk?.usedGb) }} / {{ formatGb(metrics?.disk?.totalGb) }}
              <span class="ml-2">(free {{ formatGb(metrics?.disk?.freeGb) }})</span>
            </p>
          </div>
        </section>

        <section>
          <h2 class="mb-2 text-sm font-medium text-foreground">Network</h2>
          <div class="grid gap-3 sm:grid-cols-2">
            <p class="text-sm">
              <span class="text-muted-foreground">Download</span>
              <span class="ml-2 tabular-nums font-medium">{{ formatMbps(metrics?.network?.downloadMbps) }}</span>
            </p>
            <p class="text-sm">
              <span class="text-muted-foreground">Upload</span>
              <span class="ml-2 tabular-nums font-medium">{{ formatMbps(metrics?.network?.uploadMbps) }}</span>
            </p>
          </div>
        </section>
      </CardContent>
    </Card>
  </div>
</template>
