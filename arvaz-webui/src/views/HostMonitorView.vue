<script setup lang="ts">
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'
import { Cpu, HardDrive, MemoryStick, Network } from 'lucide-vue-next'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { Line } from 'vue-chartjs'
import {
  api,
  type HostMetrics,
  type HostMetricsHistoryPoint,
  type HostMetricsHistoryRange,
} from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

const selectClass = 'rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground'

const metrics = ref<HostMetrics | null>(null)
const error = ref('')
const loading = ref(false)
const refreshSeconds = ref(2)
const refreshChoices = [1, 2, 4] as const
let refreshTimer: ReturnType<typeof setInterval> | undefined

type HistoryRange = HostMetricsHistoryRange
const historyRanges: { value: HistoryRange; label: string }[] = [
  { value: 'today', label: 'Today' },
  { value: 'yesterday', label: 'Yesterday' },
  { value: 'this_week', label: 'This week' },
  { value: 'last_week', label: 'Last week' },
  { value: 'this_month', label: 'This month' },
  { value: 'all', label: 'All time' },
]

const cpuRange = ref<HistoryRange>('today')
const memRange = ref<HistoryRange>('today')
const diskRange = ref<HistoryRange>('today')
const netRange = ref<HistoryRange>('today')

const cpuPoints = ref<HostMetricsHistoryPoint[]>([])
const memPoints = ref<HostMetricsHistoryPoint[]>([])
const diskPoints = ref<HostMetricsHistoryPoint[]>([])
const netPoints = ref<HostMetricsHistoryPoint[]>([])

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

async function loadHistory(
  range: HistoryRange,
  target: typeof cpuPoints,
) {
  try {
    const res = await api.hostMetricsHistory(range)
    target.value = res.points ?? []
  } catch {
    target.value = []
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
  void loadHistory(cpuRange.value, cpuPoints)
  void loadHistory(memRange.value, memPoints)
  void loadHistory(diskRange.value, diskPoints)
  void loadHistory(netRange.value, netPoints)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})

watch(refreshSeconds, armRefresh)
watch(cpuRange, (r) => void loadHistory(r, cpuPoints))
watch(memRange, (r) => void loadHistory(r, memPoints))
watch(diskRange, (r) => void loadHistory(r, diskPoints))
watch(netRange, (r) => void loadHistory(r, netPoints))

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
const avgCpu = computed(() => {
  const cores = metrics.value?.cpuCores
  if (!cores?.length) return 0
  return cores.reduce((a, b) => a + b, 0) / cores.length
})

const chartOpts = {
  responsive: true,
  maintainAspectRatio: false,
  animation: false as const,
  plugins: { legend: { display: false } },
  scales: {
    x: {
      ticks: { color: '#9d9d9d', maxTicksLimit: 6 },
      grid: { color: '#3c3c3c55' },
    },
    y: {
      ticks: { color: '#9d9d9d' },
      grid: { color: '#3c3c3c55' },
      beginAtZero: true,
    },
  },
}

function labels(points: HostMetricsHistoryPoint[]) {
  return points.map((p) => {
    const d = new Date(p.ts)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
  })
}

function lineData(points: HostMetricsHistoryPoint[], values: number[], color: string) {
  return {
    labels: labels(points),
    datasets: [
      {
        data: values,
        borderColor: color,
        backgroundColor: color + '33',
        fill: true,
        tension: 0.3,
        pointRadius: 0,
        borderWidth: 2,
      },
    ],
  }
}

const cpuChart = computed(() =>
  lineData(
    cpuPoints.value,
    cpuPoints.value.map((p) => p.cpuPct),
    '#007acc',
  ),
)
const memChart = computed(() =>
  lineData(
    memPoints.value,
    memPoints.value.map((p) => usagePct(p.memUsedGb, p.memTotalGb)),
    '#4ec9b0',
  ),
)
const diskChart = computed(() =>
  lineData(
    diskPoints.value,
    diskPoints.value.map((p) => usagePct(p.diskUsedGb, p.diskTotalGb)),
    '#dcdcaa',
  ),
)
const netChart = computed(() => ({
  labels: labels(netPoints.value),
  datasets: [
    {
      label: 'Down',
      data: netPoints.value.map((p) => p.netDownMbps),
      borderColor: '#569cd6',
      backgroundColor: '#569cd633',
      fill: true,
      tension: 0.3,
      pointRadius: 0,
      borderWidth: 2,
    },
    {
      label: 'Up',
      data: netPoints.value.map((p) => p.netUpMbps),
      borderColor: '#c586c0',
      backgroundColor: '#c586c033',
      fill: true,
      tension: 0.3,
      pointRadius: 0,
      borderWidth: 2,
    },
  ],
}))

const netChartOpts = {
  ...chartOpts,
  plugins: { legend: { display: true, labels: { color: '#9d9d9d' } } },
}
</script>

<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card class="server-live-card overflow-hidden border-border/80">
      <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0 border-b border-border/60 bg-[#2a2d2e]/40">
        <div>
          <CardTitle class="flex items-center gap-2 text-xl">
            <Cpu :size="20" class="text-primary" aria-hidden="true" />
            Server
          </CardTitle>
          <p class="mt-1 text-sm text-muted-foreground">
            Live CPU, memory, disk, and network · updates while this page is open
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
      <CardContent class="space-y-6 pt-6">
        <div class="grid gap-3 sm:grid-cols-4">
          <div class="rounded-lg border border-border/70 bg-muted/30 p-3">
            <p class="text-xs uppercase tracking-wide text-muted-foreground">CPU avg</p>
            <p class="mt-1 text-2xl font-semibold tabular-nums text-primary">{{ formatPct(avgCpu) }}</p>
          </div>
          <div class="rounded-lg border border-border/70 bg-muted/30 p-3">
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Memory</p>
            <p class="mt-1 text-2xl font-semibold tabular-nums">{{ formatPct(memoryPct) }}</p>
          </div>
          <div class="rounded-lg border border-border/70 bg-muted/30 p-3">
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Disk</p>
            <p class="mt-1 text-2xl font-semibold tabular-nums">{{ formatPct(diskPct) }}</p>
          </div>
          <div class="rounded-lg border border-border/70 bg-muted/30 p-3">
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Network</p>
            <p class="mt-1 text-sm tabular-nums">
              ↓ {{ formatMbps(metrics?.network?.downloadMbps) }}
              <span class="mx-1 text-muted-foreground">·</span>
              ↑ {{ formatMbps(metrics?.network?.uploadMbps) }}
            </p>
          </div>
        </div>

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
            <h2 class="mb-2 flex items-center gap-1.5 text-sm font-medium text-foreground">
              <MemoryStick :size="14" aria-hidden="true" /> Memory
            </h2>
            <div class="mb-1 h-2 overflow-hidden rounded bg-muted">
              <div class="h-full bg-[#4ec9b0] transition-[width] duration-300" :style="{ width: barWidth(memoryPct) }" />
            </div>
            <p class="text-sm tabular-nums text-muted-foreground">
              {{ formatGb(metrics?.memory?.usedGb) }} / {{ formatGb(metrics?.memory?.totalGb) }}
              <span class="ml-2">(avail {{ formatGb(metrics?.memory?.availableGb) }})</span>
            </p>
          </div>
          <div>
            <h2 class="mb-2 flex items-center gap-1.5 text-sm font-medium text-foreground">
              <HardDrive :size="14" aria-hidden="true" /> Disk
            </h2>
            <div class="mb-1 h-2 overflow-hidden rounded bg-muted">
              <div class="h-full bg-[#dcdcaa] transition-[width] duration-300" :style="{ width: barWidth(diskPct) }" />
            </div>
            <p class="text-sm tabular-nums text-muted-foreground">
              {{ formatGb(metrics?.disk?.usedGb) }} / {{ formatGb(metrics?.disk?.totalGb) }}
              <span class="ml-2">(free {{ formatGb(metrics?.disk?.freeGb) }})</span>
            </p>
          </div>
        </section>
      </CardContent>
    </Card>

    <div class="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between gap-2 space-y-0">
          <CardTitle class="flex items-center gap-2 text-base">
            <Cpu :size="16" aria-hidden="true" /> CPU history
          </CardTitle>
          <select v-model="cpuRange" :class="selectClass">
            <option v-for="r in historyRanges" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </CardHeader>
        <CardContent>
          <div class="h-48">
            <Line v-if="cpuPoints.length" :data="cpuChart" :options="chartOpts" />
            <p v-else class="text-sm text-muted-foreground">No history yet</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between gap-2 space-y-0">
          <CardTitle class="flex items-center gap-2 text-base">
            <MemoryStick :size="16" aria-hidden="true" /> Memory history
          </CardTitle>
          <select v-model="memRange" :class="selectClass">
            <option v-for="r in historyRanges" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </CardHeader>
        <CardContent>
          <div class="h-48">
            <Line v-if="memPoints.length" :data="memChart" :options="chartOpts" />
            <p v-else class="text-sm text-muted-foreground">No history yet</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between gap-2 space-y-0">
          <CardTitle class="flex items-center gap-2 text-base">
            <HardDrive :size="16" aria-hidden="true" /> Disk history
          </CardTitle>
          <select v-model="diskRange" :class="selectClass">
            <option v-for="r in historyRanges" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </CardHeader>
        <CardContent>
          <div class="h-48">
            <Line v-if="diskPoints.length" :data="diskChart" :options="chartOpts" />
            <p v-else class="text-sm text-muted-foreground">No history yet</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader class="flex flex-row items-center justify-between gap-2 space-y-0">
          <CardTitle class="flex items-center gap-2 text-base">
            <Network :size="16" aria-hidden="true" /> Network history
          </CardTitle>
          <select v-model="netRange" :class="selectClass">
            <option v-for="r in historyRanges" :key="r.value" :value="r.value">{{ r.label }}</option>
          </select>
        </CardHeader>
        <CardContent>
          <div class="h-48">
            <Line v-if="netPoints.length" :data="netChart" :options="netChartOpts" />
            <p v-else class="text-sm text-muted-foreground">No history yet</p>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
