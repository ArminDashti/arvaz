<script setup lang="ts">
import { ArrowDownToLine, Clock, Globe, HardDriveDownload, HardDriveUpload, User } from 'lucide-vue-next'
import { computed, onMounted, onUnmounted, ref } from 'vue'
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
  formatBytes,
  formatConnectedAt,
  formatMbps,
  type SortDir,
} from '@/lib/utils'

type SortKey = 'username' | 'lastIsp' | 'downloadMbps' | 'downloadBytes' | 'uploadBytes' | 'connectedAt'

const sessions = ref<SoftEtherSession[]>([])
const error = ref('')
const loading = ref(false)
const sortKey = ref<SortKey>('username')
const sortDir = ref<SortDir>('asc')
let refreshTimer: ReturnType<typeof setInterval> | undefined

async function load(silent = false) {
  if (!silent && loading.value) return
  if (!silent) loading.value = true
  try {
    const res = await api.softetherSessions()
    sessions.value = res.sessions ?? []
    error.value = res.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    if (!silent) loading.value = false
  }
}

function sortLabel(key: SortKey, label: string) {
  if (sortKey.value !== key) return label
  return `${label} ${sortDir.value === 'asc' ? '↑' : '↓'}`
}

function sortValue(s: SoftEtherSession, key: SortKey): unknown {
  switch (key) {
    case 'username':
      return s.username
    case 'lastIsp':
      return s.lastIsp || ''
    case 'downloadMbps':
      return s.downloadMbps ?? -1
    case 'downloadBytes':
      return s.downloadBytes ?? -1
    case 'uploadBytes':
      return s.uploadBytes ?? -1
    case 'connectedAt':
      return s.connectedAt || ''
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

const { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages } =
  useClientPagination(sortedSessions)

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = key === 'username' || key === 'lastIsp' ? 'asc' : 'desc'
  }
  resetPage()
}

onMounted(() => {
  void load()
  refreshTimer = setInterval(() => {
    void load(true)
  }, 1000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
      <div>
        <CardTitle>SE Online users</CardTitle>
        <p class="mt-1 text-sm text-muted-foreground">Live RX Mbps · RX/TX usage · refresh 1s</p>
      </div>
      <button type="button" class="field-button" :disabled="loading" @click="() => load()">
        {{ loading ? 'Refreshing…' : 'Refresh' }}
      </button>
    </CardHeader>
    <CardContent>
      <p v-if="error" class="mb-3 text-sm text-destructive">{{ error }}</p>
      <div style="overflow-x: auto; max-width: 100%">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 font-medium hover:underline"
                  @click="toggleSort('username')"
                >
                  <User :size="14" aria-hidden="true" />
                  {{ sortLabel('username', 'Username') }}
                </button>
              </TableHead>
              <TableHead class="min-w-64">
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 font-medium hover:underline"
                  @click="toggleSort('lastIsp')"
                >
                  <Globe :size="14" aria-hidden="true" />
                  {{ sortLabel('lastIsp', 'ISP') }}
                </button>
              </TableHead>
              <TableHead>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 font-medium hover:underline"
                  @click="toggleSort('downloadMbps')"
                >
                  <ArrowDownToLine :size="14" aria-hidden="true" />
                  {{ sortLabel('downloadMbps', 'RX') }}
                </button>
              </TableHead>
              <TableHead>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 font-medium hover:underline"
                  @click="toggleSort('downloadBytes')"
                >
                  <HardDriveDownload :size="14" aria-hidden="true" />
                  {{ sortLabel('downloadBytes', 'RX Usage') }}
                </button>
              </TableHead>
              <TableHead>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 font-medium hover:underline"
                  @click="toggleSort('uploadBytes')"
                >
                  <HardDriveUpload :size="14" aria-hidden="true" />
                  {{ sortLabel('uploadBytes', 'TX Usage') }}
                </button>
              </TableHead>
              <TableHead>
                <button
                  type="button"
                  class="inline-flex items-center gap-1.5 font-medium hover:underline"
                  @click="toggleSort('connectedAt')"
                >
                  <Clock :size="14" aria-hidden="true" />
                  {{ sortLabel('connectedAt', 'Connected at') }}
                </button>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="sortedSessions.length === 0">
              <TableCell colspan="6" class="text-muted-foreground">No online sessions</TableCell>
            </TableRow>
            <TableRow
              v-for="s in pageRows"
              :key="(s.sessionKey || '') + s.username + (s.clientIp || '') + (s.connectedAt || '')"
            >
              <TableCell>
                <RouterLink
                  v-if="s.username"
                  class="font-medium hover:underline"
                  :to="`/softether/users/${encodeURIComponent(s.username)}/logs`"
                >
                  {{ s.username }}
                </RouterLink>
                <span v-else>{{ emptyGridValue }}</span>
              </TableCell>
              <TableCell class="min-w-64">
                <IspName :name="s.lastIsp" :logo-key="s.ispLogo" :ip="s.clientIp" />
              </TableCell>
              <TableCell>{{ formatMbps(s.downloadMbps) }}</TableCell>
              <TableCell>{{ formatBytes(s.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(s.uploadBytes) }}</TableCell>
              <TableCell>
                {{
                  formatConnectedAt(
                    s.connectedAt,
                    s.sessionDurationSeconds,
                  )
                }}
              </TableCell>
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
