<script setup lang="ts">
import {
  Box,
  Check,
  ChevronDown,
  ChevronRight,
  Cpu,
  HardDrive,
  MemoryStick,
  Network,
  Play,
  Square,
  Terminal,
  Timer,
} from 'lucide-vue-next'
import { computed, onMounted, ref } from 'vue'
import { api, type DockerContainer } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatDuration, formatPercent } from '@/lib/utils'

const containers = ref<DockerContainer[]>([])
const error = ref('')
const loading = ref(false)
const actionBusy = ref('')
const collapsed = ref<Record<string, boolean>>({})

const shellOpen = ref(false)
const shellContainer = ref('')
const shellInput = ref('')
const shellLog = ref('')
const shellBusy = ref(false)

const stacks = computed(() => {
  const map = new Map<string, DockerContainer[]>()
  for (const c of containers.value) {
    const list = map.get(c.stackName) ?? []
    list.push(c)
    map.set(c.stackName, list)
  }
  return [...map.entries()].map(([stackName, items]) => ({ stackName, items }))
})

const flatRows = computed(() => {
  const rows: Array<
    | { kind: 'stack'; stackName: string; count: number }
    | { kind: 'container'; container: DockerContainer }
  > = []
  for (const stack of stacks.value) {
    rows.push({ kind: 'stack', stackName: stack.stackName, count: stack.items.length })
    if (!collapsed.value[stack.stackName]) {
      for (const c of stack.items) {
        rows.push({ kind: 'container', container: c })
      }
    }
  }
  return rows
})

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
  if (n == null || Number.isNaN(n) || n <= 0) return '—'
  return `${n.toFixed(2)} GB`
}

function toggleStack(name: string) {
  collapsed.value = { ...collapsed.value, [name]: !collapsed.value[name] }
}

async function startContainer(name: string) {
  actionBusy.value = name
  try {
    const res = await api.dockerStart(name)
    if (res.error) error.value = res.error
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Start failed'
  } finally {
    actionBusy.value = ''
  }
}

async function stopContainer(name: string) {
  if (!window.confirm(`Stop container ${name}?`)) return
  actionBusy.value = name
  try {
    const res = await api.dockerStop(name)
    if (res.error) error.value = res.error
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Stop failed'
  } finally {
    actionBusy.value = ''
  }
}

function openShell(name: string) {
  shellContainer.value = name
  shellLog.value = `$ connected to ${name}\n`
  shellInput.value = ''
  shellOpen.value = true
}

async function runShell() {
  const cmd = shellInput.value.trim()
  if (!cmd || shellBusy.value) return
  shellBusy.value = true
  shellLog.value += `$ ${cmd}\n`
  shellInput.value = ''
  try {
    const res = await api.dockerExec(shellContainer.value, cmd)
    if (res.output) shellLog.value += res.output.replace(/\n?$/, '\n')
    if (res.error) shellLog.value += `error: ${res.error}\n`
  } catch (e) {
    shellLog.value += `error: ${e instanceof Error ? e.message : 'exec failed'}\n`
  } finally {
    shellBusy.value = false
  }
}

onMounted(load)
</script>

<template>
  <Card>
    <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
      <CardTitle class="flex items-center gap-2">
        <Box :size="18" aria-hidden="true" />
        Docker
      </CardTitle>
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
                <span class="inline-flex items-center gap-1.5"><Box :size="14" aria-hidden="true" /> Stack</span>
              </TableHead>
              <TableHead>Container</TableHead>
              <TableHead>
                <span class="inline-flex items-center gap-1.5"><Cpu :size="14" aria-hidden="true" /> CPU</span>
              </TableHead>
              <TableHead>
                <span class="inline-flex items-center gap-1.5"><MemoryStick :size="14" aria-hidden="true" /> Memory</span>
              </TableHead>
              <TableHead>
                <span class="inline-flex items-center gap-1.5"><HardDrive :size="14" aria-hidden="true" /> Disk</span>
              </TableHead>
              <TableHead>
                <span class="inline-flex items-center gap-1.5"><Network :size="14" aria-hidden="true" /> Network</span>
              </TableHead>
              <TableHead>IP:port</TableHead>
              <TableHead>HAProxy/Outside</TableHead>
              <TableHead>
                <span class="inline-flex items-center gap-1.5"><Timer :size="14" aria-hidden="true" /> Uptime</span>
              </TableHead>
              <TableHead>State</TableHead>
              <TableHead>Stop/Start</TableHead>
              <TableHead>Connect</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="flatRows.length === 0">
              <TableCell colspan="12" class="text-muted-foreground">No containers</TableCell>
            </TableRow>
            <template v-for="(row, idx) in flatRows" :key="idx">
              <TableRow v-if="row.kind === 'stack'" class="bg-muted/20">
                <TableCell colspan="12">
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 font-medium hover:underline"
                    @click="toggleStack(row.stackName)"
                  >
                    <ChevronDown v-if="!collapsed[row.stackName]" :size="16" aria-hidden="true" />
                    <ChevronRight v-else :size="16" aria-hidden="true" />
                    {{ row.stackName }}
                    <span class="text-muted-foreground font-normal">({{ row.count }})</span>
                  </button>
                </TableCell>
              </TableRow>
              <TableRow v-else>
                <TableCell class="text-muted-foreground" />
                <TableCell>{{ row.container.containerName }}</TableCell>
                <TableCell>{{ formatPercent(row.container.cpuPercent) }}</TableCell>
                <TableCell>{{ formatGB(row.container.memoryGb) }}</TableCell>
                <TableCell>{{ formatGB(row.container.diskGb) }}</TableCell>
                <TableCell class="max-w-[220px] truncate" :title="row.container.network">
                  {{ row.container.network }}
                </TableCell>
                <TableCell class="max-w-[180px] truncate" :title="row.container.ipPort">
                  {{ row.container.ipPort }}
                </TableCell>
                <TableCell class="max-w-[200px] truncate" :title="row.container.haproxyUrl">
                  {{ row.container.haproxyUrl }}
                </TableCell>
                <TableCell>
                  {{
                    row.container.state === 'running'
                      ? formatDuration(row.container.uptimeSeconds)
                      : row.container.state
                  }}
                </TableCell>
                <TableCell>
                  <span class="inline-flex items-center" :title="row.container.state">
                    <Check
                      v-if="row.container.state === 'running'"
                      class="text-emerald-500"
                      :size="18"
                      aria-label="running"
                    />
                    <span v-else class="text-muted-foreground">{{ row.container.state }}</span>
                  </span>
                </TableCell>
                <TableCell>
                  <button
                    v-if="row.container.state === 'running'"
                    type="button"
                    class="field-button inline-flex items-center gap-1 px-2 py-1 text-xs"
                    :disabled="actionBusy === row.container.containerName"
                    @click="stopContainer(row.container.containerName)"
                  >
                    <Square :size="12" aria-hidden="true" /> Stop
                  </button>
                  <button
                    v-else
                    type="button"
                    class="field-button inline-flex items-center gap-1 px-2 py-1 text-xs"
                    :disabled="actionBusy === row.container.containerName"
                    @click="startContainer(row.container.containerName)"
                  >
                    <Play :size="12" aria-hidden="true" /> Start
                  </button>
                </TableCell>
                <TableCell>
                  <button
                    type="button"
                    class="field-button inline-flex items-center gap-1 px-2 py-1 text-xs"
                    @click="openShell(row.container.containerName)"
                  >
                    <Terminal :size="12" aria-hidden="true" /> Connect
                  </button>
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </div>
    </CardContent>
  </Card>

  <div
    v-if="shellOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
    @click.self="shellOpen = false"
  >
    <div class="flex w-full max-w-3xl flex-col gap-2 rounded-lg border border-border bg-card p-4 shadow-xl">
      <div class="flex items-center justify-between gap-2">
        <h2 class="text-sm font-medium">Shell · {{ shellContainer }}</h2>
        <button type="button" class="field-button px-2 py-1 text-xs" @click="shellOpen = false">Close</button>
      </div>
      <pre
        class="h-72 overflow-auto rounded border border-border bg-[#1e1e1e] p-3 font-mono text-xs text-[#d4d4d4] whitespace-pre-wrap"
      >{{ shellLog }}</pre>
      <form class="flex gap-2" @submit.prevent="runShell">
        <input
          v-model="shellInput"
          type="text"
          class="flex-1 rounded-md border border-border bg-background px-2 py-1.5 font-mono text-sm"
          placeholder="command (e.g. ls -la)"
          :disabled="shellBusy"
          autocomplete="off"
        />
        <button type="submit" class="field-button" :disabled="shellBusy || !shellInput.trim()">
          {{ shellBusy ? '…' : 'Run' }}
        </button>
      </form>
    </div>
  </div>
</template>
