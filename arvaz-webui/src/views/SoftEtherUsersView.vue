<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type SoftEtherUser } from '@/api/client'
import IspName from '@/components/IspName.vue'
import TablePagination from '@/components/TablePagination.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useClientPagination } from '@/composables/useClientPagination'
import { compareSortValues, emptyGridValue, formatBytes, formatDuration, formatLastLogin, softetherIpLogsPath, type SortDir } from '@/lib/utils'

type SortKey =
  | 'username'
  | 'downloadBytes'
  | 'uploadBytes'
  | 'lastIp'
  | 'lastIsp'
  | 'lastLogin'
  | 'trafficYesterdayBytes'
  | 'trafficWeekBytes'
  | 'trafficMonthBytes'
  | 'trafficTotalBytes'

const users = ref<SoftEtherUser[]>([])
const error = ref('')
const loading = ref(false)
const sortKey = ref<SortKey>('username')
const sortDir = ref<SortDir>('asc')

async function load() {
  loading.value = true
  try {
    const res = await api.softetherUsers()
    users.value = res.users ?? []
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
    sortDir.value = key === 'username' || key === 'lastIp' || key === 'lastIsp' ? 'asc' : 'desc'
  }
  resetPage()
}

function sortLabel(key: SortKey, label: string) {
  if (sortKey.value !== key) return label
  return `${label} ${sortDir.value === 'asc' ? '↑' : '↓'}`
}

function sortValue(u: SoftEtherUser, key: SortKey): unknown {
  switch (key) {
    case 'username':
      return u.username
    case 'downloadBytes':
      return u.downloadBytes ?? 0
    case 'uploadBytes':
      return u.uploadBytes ?? 0
    case 'lastIp':
      return u.lastIp || ''
    case 'lastIsp':
      return u.lastIsp || ''
    case 'lastLogin':
      return u.lastLogin || ''
    case 'trafficYesterdayBytes':
      return u.trafficYesterdayBytes ?? 0
    case 'trafficWeekBytes':
      return u.trafficWeekBytes ?? 0
    case 'trafficMonthBytes':
      return u.trafficMonthBytes ?? 0
    case 'trafficTotalBytes':
      return u.trafficTotalBytes ?? (u.downloadBytes ?? 0) + (u.uploadBytes ?? 0)
    default: {
      const _exhaustive: never = key
      return _exhaustive
    }
  }
}

const sortedUsers = computed(() => {
  const rows = [...users.value]
  const key = sortKey.value
  const dir = sortDir.value
  rows.sort((a, b) => compareSortValues(sortValue(a, key), sortValue(b, key), dir))
  return rows
})

const { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages } = useClientPagination(sortedUsers)

onMounted(load)
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
      <CardTitle>SE Users</CardTitle>
      <button type="button" class="field-button" :disabled="loading" @click="load">
        {{ loading ? 'Refreshing…' : 'Refresh' }}
      </button>
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
                  {{ sortLabel('downloadBytes', 'Total Download') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('uploadBytes')">
                  {{ sortLabel('uploadBytes', 'Total Upload') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('lastIp')">
                  {{ sortLabel('lastIp', 'Last IP') }}
                </button>
              </TableHead>
              <TableHead class="min-w-64">
                <button type="button" class="font-medium hover:underline" @click="toggleSort('lastIsp')">
                  {{ sortLabel('lastIsp', 'ISP') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('lastLogin')">
                  {{ sortLabel('lastLogin', 'Last login') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('trafficYesterdayBytes')">
                  {{ sortLabel('trafficYesterdayBytes', 'Yesterday') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('trafficWeekBytes')">
                  {{ sortLabel('trafficWeekBytes', 'Week') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('trafficMonthBytes')">
                  {{ sortLabel('trafficMonthBytes', 'Month') }}
                </button>
              </TableHead>
              <TableHead>
                <button type="button" class="font-medium hover:underline" @click="toggleSort('trafficTotalBytes')">
                  {{ sortLabel('trafficTotalBytes', 'Total') }}
                </button>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="sortedUsers.length === 0">
              <TableCell colspan="10" class="text-muted-foreground">No users</TableCell>
            </TableRow>
            <TableRow v-for="u in pageRows" :key="u.username">
              <TableCell>
                <RouterLink class="font-medium hover:underline" :to="`/softether/users/${encodeURIComponent(u.username)}/logs`">
                  {{ u.username }}
                </RouterLink>
              </TableCell>
              <TableCell>{{ formatBytes(u.downloadBytes) }}</TableCell>
              <TableCell>{{ formatBytes(u.uploadBytes) }}</TableCell>
              <TableCell>
                <RouterLink
                  v-if="u.lastIp"
                  class="font-medium hover:underline"
                  :to="softetherIpLogsPath(u.lastIp)"
                >
                  {{ u.lastIp }}
                </RouterLink>
                <span v-else>{{ emptyGridValue }}</span>
              </TableCell>
              <TableCell class="min-w-64">
                <IspName :name="u.lastIsp" :logo-key="u.ispLogo" />
              </TableCell>
              <TableCell>{{ formatLastLogin(u.lastLogin) }}</TableCell>
              <TableCell>{{ formatBytes(u.trafficYesterdayBytes ?? 0) }} ({{ formatDuration(u.trafficYesterdaySeconds ?? 0) }})</TableCell>
              <TableCell>{{ formatBytes(u.trafficWeekBytes ?? 0) }} ({{ formatDuration(u.trafficWeekSeconds ?? 0) }})</TableCell>
              <TableCell>{{ formatBytes(u.trafficMonthBytes ?? 0) }} ({{ formatDuration(u.trafficMonthSeconds ?? 0) }})</TableCell>
              <TableCell>{{ formatBytes(u.trafficTotalBytes ?? (u.downloadBytes ?? 0) + (u.uploadBytes ?? 0)) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <TablePagination
        :page="page"
        :total-pages="totalPages"
        :total-rows="sortedUsers.length"
        :page-size="pageSize"
        @prev="prevPage"
        @next="nextPage"
      />
    </CardContent>
  </Card>
</template>
