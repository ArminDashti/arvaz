<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { api, type SoftEtherSessionLog } from '@/api/client'
import IspName from '@/components/IspName.vue'
import TablePagination from '@/components/TablePagination.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useClientPagination } from '@/composables/useClientPagination'
import { compareSortValues, emptyGridValue, formatBytes, formatDuration, formatLastLogin, resolveDurationSeconds, softetherIpLogsPath, type SortDir } from '@/lib/utils'

type SortKey = 'downloadBytes' | 'uploadBytes' | 'ip' | 'isp' | 'durationSeconds' | 'connectedAt' | 'disconnectedAt'

const route = useRoute()
const username = computed(() => String(route.params.username || ''))
const sessions = ref<SoftEtherSessionLog[]>([])
const error = ref('')
const loading = ref(false)
const sortKey = ref<SortKey>('connectedAt')
const sortDir = ref<SortDir>('desc')

async function load() {
  loading.value = true
  try {
    const res = await api.softetherUserSessions(username.value)
    sessions.value = res.sessions ?? []
    error.value = res.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = key === 'ip' || key === 'isp' ? 'asc' : 'desc'
  }
  resetPage()
}

function sortLabel(key: SortKey, label: string) {
  if (sortKey.value !== key) return label
  return `${label} ${sortDir.value === 'asc' ? '↑' : '↓'}`
}

function sortValue(s: SoftEtherSessionLog, key: SortKey): unknown {
  switch (key) {
    case 'downloadBytes':
      return s.downloadBytes ?? 0
    case 'uploadBytes':
      return s.uploadBytes ?? 0
    case 'ip':
      return s.ip || ''
    case 'isp':
      return s.isp || ''
    case 'durationSeconds':
      return sessionTimeUsageSeconds(s)
    case 'connectedAt':
      return s.connectedAt || ''
    case 'disconnectedAt':
      return s.disconnectedAt || ''
    default: {
      const _exhaustive: never = key
      return _exhaustive
    }
  }
}

function sessionTimeUsageSeconds(s: SoftEtherSessionLog): number {
  return resolveDurationSeconds({
    durationSeconds: s.durationSeconds,
    connectedAt: s.connectedAt,
    disconnectedAt: s.disconnectedAt,
  })
}

const totalTimeUsageSeconds = computed(() =>
  sessions.value.reduce((sum, row) => sum + sessionTimeUsageSeconds(row), 0),
)

const sortedSessions = computed(() => {
  const rows = [...sessions.value]
  const key = sortKey.value
  const dir = sortDir.value
  rows.sort((a, b) => compareSortValues(sortValue(a, key), sortValue(b, key), dir))
  return rows
})

const { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages } = useClientPagination(sortedSessions)

watch(username, () => {
  resetPage()
  load()
})
onMounted(load)
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
      <div class="min-w-0">
        <CardTitle>Connection logs — {{ username }}</CardTitle>
        <p class="mt-1 text-sm text-muted-foreground">Time usage {{ formatDuration(totalTimeUsageSeconds) }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <RouterLink to="/softether/users" class="field-button">SE Users</RouterLink>
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
                <button type="button" class="font-medium hover:underline" @click="toggleSort('ip')">
                  {{ sortLabel('ip', 'IP') }}
                </button>
              </TableHead>
              <TableHead class="min-w-64">
                <button type="button" class="font-medium hover:underline" @click="toggleSort('isp')">
                  {{ sortLabel('isp', 'ISP') }}
                </button>
              </TableHead>
              <TableHead class="min-w-28">
                <button type="button" class="font-medium hover:underline" @click="toggleSort('durationSeconds')">
                  {{ sortLabel('durationSeconds', 'Time usage') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('connectedAt')">
                  {{ sortLabel('connectedAt', 'Connected at') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('disconnectedAt')">
                  {{ sortLabel('disconnectedAt', 'Disconnected at') }}
                </button>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="sortedSessions.length === 0">
              <TableCell colspan="7" class="text-muted-foreground">No connection logs</TableCell>
            </TableRow>
            <TableRow v-for="(s, i) in pageRows" :key="(s.connectedAt || '') + (s.ip || '') + i">
              <TableCell>{{ formatBytes(s.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(s.uploadBytes) }}</TableCell>
              <TableCell>
                <RouterLink
                  v-if="s.ip"
                  class="font-medium hover:underline"
                  :to="softetherIpLogsPath(s.ip)"
                >
                  {{ s.ip }}
                </RouterLink>
                <span v-else>{{ emptyGridValue }}</span>
              </TableCell>
              <TableCell class="min-w-64">
                <IspName :name="s.isp" :logo-key="s.ispLogo" />
              </TableCell>
              <TableCell class="min-w-28">{{ formatDuration(sessionTimeUsageSeconds(s)) }}</TableCell>
              <TableCell>{{ formatLastLogin(s.connectedAt) }}</TableCell>
              <TableCell>{{ s.disconnectedAt ? formatLastLogin(s.disconnectedAt) : '—' }}</TableCell>
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
