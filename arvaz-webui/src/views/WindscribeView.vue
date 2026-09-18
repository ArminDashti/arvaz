<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  api,
  type WindscribeLocation,
  type WindscribePingResult,
  type WindscribeSpeedtestResult,
  type WindscribeStatus,
} from '@/api/client'
import CountryFlag from '@/components/CountryFlag.vue'
import TablePagination from '@/components/TablePagination.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useClientPagination } from '@/composables/useClientPagination'

type ServerMode = 'country' | 'city' | 'all'

const status = ref<WindscribeStatus | null>(null)
const locations = ref<WindscribeLocation[]>([])
const error = ref('')
const loading = ref(false)
const busy = ref('')
const pingTarget = ref('1.1.1.1')
const pingCount = ref(4)
const pingResult = ref<WindscribePingResult | null>(null)
const speed = ref<WindscribeSpeedtestResult | null>(null)
const serverMode = ref<ServerMode>('country')
const filter = ref('')

const pingCounts = [2, 4, 8, 16] as const

const statusCountryCode = computed(() => {
  const active = locations.value.find((l) => l.active)
  if (active?.countryCode) return active.countryCode
  return locations.value.find((l) => l.region === status.value?.country)?.countryCode
})

const filteredLocations = computed(() => {
  const q = filter.value.trim().toLowerCase()
  let rows = locations.value
  if (serverMode.value === 'country') {
    const seen = new Set<string>()
    rows = rows.filter((l) => {
      const key = l.region
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
  } else if (serverMode.value === 'city') {
    const seen = new Set<string>()
    rows = rows.filter((l) => {
      const key = `${l.region}|${l.city}`
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
  }
  if (!q) return rows
  return rows.filter((l) =>
    [l.region, l.city, l.nickname, l.label, l.countryCode].join(' ').toLowerCase().includes(q),
  )
})

const { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages } =
  useClientPagination(filteredLocations)

watch([filter, serverMode], resetPage)

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const [st, locs] = await Promise.all([api.windscribeStatus(), api.windscribeLocations()])
    if (st.error) error.value = st.error
    if (locs.error) error.value = locs.error || error.value
    status.value = st.status ?? null
    locations.value = locs.locations ?? []
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load Windscribe'
  } finally {
    loading.value = false
  }
}

function connectTarget(loc: WindscribeLocation): string {
  if (serverMode.value === 'country') return loc.region
  if (serverMode.value === 'city') return loc.city
  return loc.nickname || loc.label
}

async function connectLocation(loc: WindscribeLocation) {
  const target = connectTarget(loc)
  const label =
    serverMode.value === 'country'
      ? loc.region
      : serverMode.value === 'city'
        ? `${loc.city}, ${loc.region}`
        : loc.label
  const ok = window.confirm(
    `Connect Windscribe to ${label}? SoftEther VPN users may briefly stall.`,
  )
  if (!ok) return
  busy.value = `connect:${loc.label}`
  error.value = ''
  try {
    const res = await api.windscribeConnect(target)
    if (res.error) error.value = res.error
    await refresh()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Connect failed'
  } finally {
    busy.value = ''
  }
}

async function disconnect() {
  const ok = window.confirm(
    'Disconnect Windscribe? SoftEther VPN users may lose egress until you reconnect.',
  )
  if (!ok) return
  busy.value = 'disconnect'
  error.value = ''
  try {
    const res = await api.windscribeDisconnect()
    if (res.error) error.value = res.error
    await refresh()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Disconnect failed'
  } finally {
    busy.value = ''
  }
}

async function runPing() {
  busy.value = 'ping'
  pingResult.value = null
  const count = pingCounts.includes(pingCount.value as (typeof pingCounts)[number])
    ? pingCount.value
    : 4
  pingCount.value = count
  try {
    const res = await api.windscribePing(pingTarget.value, count)
    if (res.error) error.value = res.error
    pingResult.value = res.result ?? null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Ping failed'
  } finally {
    busy.value = ''
  }
}

async function runSpeed(mode: 'single' | 'parallel') {
  busy.value = mode
  speed.value = null
  error.value = ''
  try {
    const res = await api.windscribeSpeedtest(mode)
    if (res.error) error.value = res.error
    speed.value = res.result ?? null
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Speedtest failed'
  } finally {
    busy.value = ''
  }
}

function formatMbps(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n >= 10 ? n.toFixed(0) : n.toFixed(1)} Mbps`
}

onMounted(refresh)
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-end gap-3">
      <button type="button" class="field-button" :disabled="loading || !!busy" @click="refresh">
        {{ loading ? 'Refreshing…' : 'Refresh' }}
      </button>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <div class="grid gap-4 lg:grid-cols-[minmax(0,22rem)_1fr] lg:items-stretch">
      <div class="flex min-w-0 flex-col gap-4">
        <Card>
          <CardContent class="space-y-3 pt-4">
            <p
              class="text-2xl font-semibold"
              :class="status?.connected ? 'text-emerald-500' : 'text-red-500'"
            >
              {{ status?.connected ? 'Connected' : 'Disconnected' }}
            </p>
            <CountryFlag
              size="lg"
              :country-code="statusCountryCode"
              :country-name="status?.country"
            />
            <div class="space-y-1 text-sm">
              <p>
                <span class="text-muted-foreground">Public IP:</span>
                {{ status?.publicIp || '—' }}
              </p>
              <p>
                <span class="text-muted-foreground">Location:</span>
                {{ status?.location || '—' }}
              </p>
              <p>
                <span class="text-muted-foreground">Protocol:</span>
                {{ status?.protocol || '—' }}
              </p>
            </div>
            <button
              v-if="status?.connected"
              type="button"
              class="field-button w-full"
              :disabled="!!busy"
              @click="disconnect"
            >
              {{ busy === 'disconnect' ? 'Disconnecting…' : 'Disconnect' }}
            </button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Ping</CardTitle></CardHeader>
          <CardContent class="space-y-3">
            <div class="flex min-w-0 flex-wrap items-center gap-2">
              <input
                v-model="pingTarget"
                class="min-w-0 flex-1 rounded-md border border-border bg-background px-3 py-2 text-sm"
                placeholder="Endpoint"
              />
              <select
                v-model.number="pingCount"
                class="w-20 shrink-0 rounded-md border border-border bg-background px-2 py-2 text-sm"
                title="Packet count"
              >
                <option v-for="n in pingCounts" :key="n" :value="n">{{ n }}</option>
              </select>
              <button type="button" class="field-button shrink-0" :disabled="!!busy" @click="runPing">
                {{ busy === 'ping' ? 'Pinging…' : 'Run ping' }}
              </button>
            </div>
            <div v-if="pingResult" class="grid grid-cols-2 gap-3 text-sm">
              <div>
                <p class="text-xs uppercase tracking-wide text-muted-foreground">Packet loss</p>
                <p class="mt-1 text-xl font-semibold">{{ pingResult.packetLossPercent }}%</p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-muted-foreground">Average ping</p>
                <p class="mt-1 text-xl font-semibold">{{ pingResult.avgMs.toFixed(1) }} ms</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="flex flex-row items-center justify-between gap-3 space-y-0">
            <CardTitle>Speedtest (Ookla)</CardTitle>
            <div class="flex flex-wrap gap-2">
              <button type="button" class="field-button" :disabled="!!busy" @click="runSpeed('single')">
                {{ busy === 'single' ? 'Running…' : 'Single' }}
              </button>
              <button type="button" class="field-button" :disabled="!!busy" @click="runSpeed('parallel')">
                {{ busy === 'parallel' ? 'Running…' : 'Parallel' }}
              </button>
            </div>
          </CardHeader>
          <CardContent>
            <div v-if="speed?.parsedOk" class="grid gap-4 sm:grid-cols-3">
              <div>
                <p class="text-xs uppercase tracking-wide text-muted-foreground">Download</p>
                <p class="mt-1 text-3xl font-semibold text-emerald-500">{{ formatMbps(speed.downloadMbps) }}</p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-muted-foreground">Upload</p>
                <p class="mt-1 text-3xl font-semibold text-red-500">{{ formatMbps(speed.uploadMbps) }}</p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-muted-foreground">Latency</p>
                <p class="mt-1 text-3xl font-semibold">
                  {{ speed.latencyMs != null ? `${speed.latencyMs.toFixed(0)} ms` : '—' }}
                </p>
              </div>
            </div>
            <pre v-else-if="speed?.raw" class="max-h-40 overflow-auto rounded-md bg-muted/40 p-3 text-xs">{{ speed.raw }}</pre>
            <p v-else class="text-sm text-muted-foreground">Run a speedtest to see Download, Upload, and Latency.</p>
          </CardContent>
        </Card>
      </div>

      <Card class="flex min-h-0 min-w-0 flex-col">
        <CardHeader class="flex flex-col gap-3 space-y-0 sm:flex-row sm:items-center sm:justify-between">
          <CardTitle>Locations</CardTitle>
          <div class="flex flex-wrap items-center gap-2">
            <div class="flex rounded-md border border-border p-0.5 text-sm">
              <button
                type="button"
                class="rounded px-3 py-1.5"
                :class="serverMode === 'country' ? 'bg-muted font-medium' : 'text-muted-foreground'"
                @click="serverMode = 'country'"
              >
                Region
              </button>
              <button
                type="button"
                class="rounded px-3 py-1.5"
                :class="serverMode === 'city' ? 'bg-muted font-medium' : 'text-muted-foreground'"
                @click="serverMode = 'city'"
              >
                City
              </button>
              <button
                type="button"
                class="rounded px-3 py-1.5"
                :class="serverMode === 'all' ? 'bg-muted font-medium' : 'text-muted-foreground'"
                @click="serverMode = 'all'"
              >
                All
              </button>
            </div>
            <input
              v-model="filter"
              class="max-w-xs rounded-md border border-border bg-background px-3 py-2 text-sm"
              placeholder="Filter…"
            />
          </div>
        </CardHeader>
        <CardContent class="flex min-h-0 flex-1 flex-col">
          <div class="min-h-0 flex-1 overflow-auto" style="max-width: 100%;">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Region</TableHead>
                  <TableHead v-if="serverMode !== 'country'">City</TableHead>
                  <TableHead v-if="serverMode === 'all'">Nickname</TableHead>
                  <TableHead>Active</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-if="filteredLocations.length === 0">
                  <TableCell
                    :colspan="serverMode === 'all' ? 5 : serverMode === 'city' ? 4 : 3"
                    class="text-muted-foreground"
                  >
                    No locations
                  </TableCell>
                </TableRow>
                <TableRow v-for="loc in pageRows" :key="loc.label">
                  <TableCell>
                    <span class="inline-flex items-center gap-2">
                      <CountryFlag :country-code="loc.countryCode" :country-name="loc.region" />
                      {{ loc.region }}
                    </span>
                  </TableCell>
                  <TableCell v-if="serverMode !== 'country'">{{ loc.city }}</TableCell>
                  <TableCell v-if="serverMode === 'all'">{{ loc.nickname }}</TableCell>
                  <TableCell>{{ loc.active ? 'yes' : '' }}</TableCell>
                  <TableCell>
                    <button
                      type="button"
                      class="field-button"
                      :disabled="!!busy"
                      @click="connectLocation(loc)"
                    >
                      {{ busy === `connect:${loc.label}` ? 'Connecting…' : 'Connect' }}
                    </button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <TablePagination
            :page="page"
            :total-pages="totalPages"
            :total-rows="filteredLocations.length"
            :page-size="pageSize"
            @prev="prevPage"
            @next="nextPage"
          />
        </CardContent>
      </Card>
    </div>
  </div>
</template>
