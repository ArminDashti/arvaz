<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type SoftEtherSession } from '@/api/client'
import IspName from '@/components/IspName.vue'
import TablePagination from '@/components/TablePagination.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useClientPagination } from '@/composables/useClientPagination'
import {
  compareSortValues,
  emptyGridValue,
  formatBps,
  formatBytes,
  formatClockHHMM,
  formatDuration,
  softetherIpLogsPath,
  type SortDir,
} from '@/lib/utils'

type SortKey =
  | 'username'
  | 'downloadBytes'
  | 'uploadBytes'
  | 'bwBps'
  | 'clientIp'
  | 'lastIsp'
  | 'connectedAt'
  | 'sessionDurationSeconds'

type SessionRow = SoftEtherSession & { bwBps: number | null }

type TrafficSnapshot = {
  totalBytes: number
  atMs: number
}

const sessions = ref<SessionRow[]>([])
const error = ref('')
const loading = ref(false)
const sortKey = ref<SortKey>('username')
const sortDir = ref<SortDir>('asc')
const refreshMinutes = ref(4)
const refreshChoices = [2, 4, 8, 16] as const
let refreshTimer: ReturnType<typeof setInterval> | undefined
const priorTraffic = new Map<string, TrafficSnapshot>()
let lastSampleAtMs = 0

function sessionIdentity(s: SoftEtherSession): string {
  if (s.sessionKey) return s.sessionKey
  return `${s.username}|${s.clientIp || ''}`
}

function applyBandwidth(next: SoftEtherSession[]): SessionRow[] {
  const nowMs = Date.now()
  const deltaMs = lastSampleAtMs > 0 ? nowMs - lastSampleAtMs : 0
  const nextPrior = new Map<string, TrafficSnapshot>()
  const rows: SessionRow[] = next.map((s) => {
    const key = sessionIdentity(s)
    const totalBytes = (s.downloadBytes ?? 0) + (s.uploadBytes ?? 0)
    nextPrior.set(key, { totalBytes, atMs: nowMs })
    const prev = priorTraffic.get(key)
    let bwBps: number | null = null
    if (prev && deltaMs > 0) {
      const deltaBytes = totalBytes - prev.totalBytes
      if (deltaBytes >= 0) {
        bwBps = deltaBytes / (deltaMs / 1000)
      }
    }
    return { ...s, bwBps }
  })
  priorTraffic.clear()
  for (const [k, v] of nextPrior) priorTraffic.set(k, v)
  lastSampleAtMs = nowMs
  return rows
}

async function load() {
  if (loading.value) return
  loading.value = true
  try {
    const res = await api.softetherSessions()
    sessions.value = applyBandwidth(res.sessions ?? [])
    error.value = res.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

function sortLabel(key: SortKey, label: string) {
  if (sortKey.value !== key) return label
  return `${label} ${sortDir.value === 'asc' ? '↑' : '↓'}`
}

function sortValue(s: SessionRow, key: SortKey): unknown {
  switch (key) {
    case 'username':
      return s.username
    case 'downloadBytes':
      return s.downloadBytes ?? 0
    case 'uploadBytes':
      return s.uploadBytes ?? 0
    case 'bwBps':
      return s.bwBps ?? -1
    case 'clientIp':
      return s.clientIp || ''
    case 'lastIsp':
      return s.lastIsp || ''
    case 'connectedAt':
      return s.connectedAt || ''
    case 'sessionDurationSeconds':
      return s.sessionDurationSeconds ?? 0
    default: {
      const _exhaustive: never = key
      return _exhaustive
    }
  }
}

const sortedSessions = computed(() => {
  const rows = [...sessions.value]
  const key = sortKey.value
  const dir = sortDir.value
  rows.sort((a, b) => compareSortValues(sortValue(a, key), sortValue(b, key), dir))
  return rows
})

const { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages } = useClientPagination(sortedSessions)

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = key === 'username' || key === 'clientIp' || key === 'lastIsp' ? 'asc' : 'desc'
  }
  resetPage()
}

onMounted(() => {
  void load()
  armRefresh()
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})

function armRefresh() {
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = setInterval(() => {
    void load()
  }, refreshMinutes.value * 60_000)
}

watch(refreshMinutes, armRefresh)
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
      <CardTitle>SE Online users</CardTitle>
      <div class="flex flex-wrap items-center gap-2">
        <label class="flex items-center gap-2 text-sm text-muted-foreground">
          Auto-refresh
          <select
            v-model.number="refreshMinutes"
            class="rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground"
          >
            <option v-for="n in refreshChoices" :key="n" :value="n">{{ n }}m</option>
          </select>
        </label>
        <button type="button" class="field-button" :disabled="loading" @click="load">
          {{ loading ? 'Refreshing…' : 'Refresh' }}
        </button>
      </div>
    </CardHeader>
    <CardContent>
      <p v-if="error" class="mb-3 text-sm text-destructive">{{ error }}</p>
      <div style="overflow-x: auto; max-width: 100%;">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('username')">
                  {{ sortLabel('username', 'Username') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('downloadBytes')">
                  {{ sortLabel('downloadBytes', 'Download') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('uploadBytes')">
                  {{ sortLabel('uploadBytes', 'Upload') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('bwBps')">
                  {{ sortLabel('bwBps', 'BW') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('clientIp')">
                  {{ sortLabel('clientIp', 'IP') }}
                </button>
              </TableHead>
              <TableHead class="min-w-64">
                <button type="button" class="font-medium hover:underline" @click="toggleSort('lastIsp')">
                  {{ sortLabel('lastIsp', 'ISP') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('connectedAt')">
                  {{ sortLabel('connectedAt', 'Connected at') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('sessionDurationSeconds')">
                  {{ sortLabel('sessionDurationSeconds', 'Uptime') }}
                </button>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="sortedSessions.length === 0">
              <TableCell colspan="8" class="text-muted-foreground">No online sessions</TableCell>
            </TableRow>
            <TableRow
              v-for="s in pageRows"
              :key="(s.sessionKey || '') + s.username + (s.clientIp || '') + (s.connectedAt || '')"
            >
              <TableCell>
                <RouterLink class="font-medium hover:underline" :to="`/softether/users/${encodeURIComponent(s.username)}/logs`">
                  {{ s.username }}
                </RouterLink>
              </TableCell>
              <TableCell>{{ formatBytes(s.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(s.uploadBytes) }}</TableCell>
              <TableCell>{{ formatBps(s.bwBps) }}</TableCell>
              <TableCell>
                <RouterLink
                  v-if="s.clientIp"
                  class="font-medium hover:underline"
                  :to="softetherIpLogsPath(s.clientIp)"
                >
                  {{ s.clientIp }}
                </RouterLink>
                <span v-else>{{ emptyGridValue }}</span>
              </TableCell>
              <TableCell class="min-w-64">
                <IspName :name="s.lastIsp" :logo-key="s.ispLogo" />
              </TableCell>
              <TableCell>{{ formatClockHHMM(s.connectedAt) }}</TableCell>
              <TableCell>{{ formatDuration(s.sessionDurationSeconds) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <TablePagination
        :page="page"
        :total-pages="totalPages"
        :total-rows="sortedSessions.length"
        :page-size="pageSize"
        @prev="prevPage"
        @next="nextPage"
      />
    </CardContent>
  </Card>
</template>
