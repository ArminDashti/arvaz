<script setup lang="ts">
import { Check } from 'lucide-vue-next'
import { computed, onMounted, ref } from 'vue'
import { api, type DockerContainer } from '@/api/client'
import TablePagination from '@/components/TablePagination.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useClientPagination } from '@/composables/useClientPagination'
import { formatDuration, formatPercent } from '@/lib/utils'

const containers = ref<DockerContainer[]>([])
const error = ref('')
const loading = ref(false)

const rows = computed(() => {
  let lastStack = ''
  return containers.value.map((c) => {
    const showStack = c.stackName !== lastStack
    if (showStack) lastStack = c.stackName
    return { ...c, stackDisplay: showStack ? c.stackName : '' }
  })
})

const { page, pageRows, pageSize, prevPage, nextPage, totalPages } = useClientPagination(rows)

async function load() {
  loading.value = true
  try {
    const res = await api.dockerContainers()
    containers.value = res.containers ?? []
    error.value = res.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
}

function formatGB(n: number | undefined | null): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n.toFixed(2)} GB`
}

onMounted(load)
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
      <CardTitle>Docker</CardTitle>
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
              <TableHead>Stack</TableHead>
              <TableHead>Container</TableHead>
              <TableHead>CPU</TableHead>
              <TableHead>Memory</TableHead>
              <TableHead>Disk</TableHead>
              <TableHead>Network</TableHead>
              <TableHead>IP:port</TableHead>
              <TableHead>HAProxy</TableHead>
              <TableHead>Uptime</TableHead>
              <TableHead>State</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="rows.length === 0">
              <TableCell colspan="10" class="text-muted-foreground">No containers</TableCell>
            </TableRow>
            <TableRow v-for="c in pageRows" :key="c.containerName">
              <TableCell class="font-medium">{{ c.stackDisplay }}</TableCell>
              <TableCell>{{ c.containerName }}</TableCell>
              <TableCell>{{ formatPercent(c.cpuPercent) }}</TableCell>
              <TableCell>{{ formatGB(c.memoryGb) }}</TableCell>
              <TableCell>{{ formatGB(c.diskGb) }}</TableCell>
              <TableCell class="max-w-[220px] truncate" :title="c.network">{{ c.network }}</TableCell>
              <TableCell class="max-w-[180px] truncate" :title="c.ipPort">{{ c.ipPort }}</TableCell>
              <TableCell class="max-w-[200px] truncate" :title="c.haproxyUrl">{{ c.haproxyUrl }}</TableCell>
              <TableCell>{{ c.state === 'running' ? formatDuration(c.uptimeSeconds) : c.state }}</TableCell>
              <TableCell>
                <span class="inline-flex items-center" :title="c.state">
                  <Check v-if="c.state === 'running'" class="text-emerald-500" :size="18" aria-label="running" />
                  <span v-else class="text-muted-foreground">{{ c.state }}</span>
                </span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
      <TablePagination
        :page="page"
        :total-pages="totalPages"
        :total-rows="rows.length"
        :page-size="pageSize"
        @prev="prevPage"
        @next="nextPage"
      />
    </CardContent>
  </Card>
</template>
