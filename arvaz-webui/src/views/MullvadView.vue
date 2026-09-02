<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api, type MullvadPingResult, type MullvadRelay, type MullvadStatus, type MullvadSpeedtestResult } from '@/api/client'
import CountryFlag from '@/components/CountryFlag.vue'
import TablePagination from '@/components/TablePagination.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useClientPagination } from '@/composables/useClientPagination'
import { countryCodeFromRelay } from '@/lib/flag'

type ServerMode = 'country' | 'city' | 'all'

const status = ref<MullvadStatus | null>(null)
const relays = ref<MullvadRelay[]>([])
const error = ref('')
const loading = ref(false)
const busy = ref('')
const pingTarget = ref('1.1.1.1')
const pingCount = ref(4)
const pingResult = ref<MullvadPingResult | null>(null)
const speed = ref<MullvadSpeedtestResult | null>(null)
const antiMode = ref('lwo')
const quantumResistant = ref(true)
const daita = ref(false)
const serverMode = ref<ServerMode>('country')
const filter = ref('')

const antiModes = ['auto', 'off', 'wireguard-port', 'udp2tcp', 'shadowsocks', 'quic', 'lwo']
const pingCounts = [2, 4, 8, 16] as const

const statusCountryCode = computed(() => {
  const fromRelay = countryCodeFromRelay(status.value?.relay)
  if (fromRelay) return fromRelay
  return (
    relays.value.find((r) => r.active)?.countryCode
    || relays.value.find((r) => r.country === status.value?.country)?.countryCode
  )
})

const filteredRelays = computed(() => {
  const q = filter.value.trim().toLowerCase()
  let rows = relays.value
  if (serverMode.value === 'country') {
    const seen = new Set<string>()
    rows = rows.filter((r) => {
      const key = r.countryCode
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
  } else if (serverMode.value === 'city') {
    const seen = new Set<string>()
    rows = rows.filter((r) => {
      const key = `${r.countryCode}|${r.cityCode}`
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
  }
  if (!q) return rows
  return rows.filter((r) =>
    [r.country, r.countryCode, r.city, r.cityCode, r.hostname, r.ipv4].join(' ').toLowerCase().includes(q),
  )
})

const { page, pageRows, pageSize, prevPage, nextPage, resetPage, totalPages } = useClientPagination(filteredRelays)

watch([filter, serverMode], resetPage)

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const [st, rl] = await Promise.all([api.mullvadStatus(), api.mullvadRelays()])
    if (st.error) error.value = st.error
    if (rl.error) error.value = rl.error || error.value
    status.value = st.status ?? null
    relays.value = rl.relays ?? []
    if (st.status?.antiCensorship) antiMode.value = st.status.antiCensorship
    if (st.status) {
      quantumResistant.value = !!st.status.quantumResistant
      daita.value = !!st.status.daita
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load Mullvad'
  } finally {
    loading.value = false
  }
}

async function connectRelay(r: MullvadRelay) {
  const label =
    serverMode.value === 'country'
      ? `${r.country} (${r.countryCode})`
      : serverMode.value === 'city'
        ? `${r.city}, ${r.country}`
        : `${r.hostname} (${r.city}, ${r.country})`
  const ok = window.confirm(
    `Connect Mullvad to ${label}? SoftEther VPN users may briefly stall.`,
  )
  if (!ok) return
  busy.value = `connect:${r.hostname}`
  error.value = ''
  try {
    const country = r.countryCode
    const city = serverMode.value === 'country' ? undefined : r.cityCode
    const host = serverMode.value === 'all' ? r.hostname : undefined
    const res = await api.mullvadSetRelay(country, city, host)
    if (res.error) error.value = res.error
    await refresh()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Connect failed'
  } finally {
    busy.value = ''
  }
}

async function applyAnti() {
  const ok = window.confirm(
    'Changing anti-censorship may briefly stall SoftEther VPN users. Continue?',
  )
  if (!ok) return
  busy.value = 'anti'
  error.value = ''
  try {
    const res = await api.mullvadSetAnti(antiMode.value)
    if (res.error) error.value = res.error
    await refresh()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Anti-censorship change failed'
  } finally {
    busy.value = ''
  }
}

async function applyTunnel() {
  const ok = window.confirm(
    'Changing Quantum Resistance or DAITA may briefly stall SoftEther VPN users. Continue?',
  )
  if (!ok) return
  busy.value = 'tunnel'
  error.value = ''
  try {
    const res = await api.mullvadSetTunnel({
      quantumResistant: quantumResistant.value,
      daita: daita.value,
    })
    if (res.error) error.value = res.error
    await refresh()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Tunnel settings change failed'
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
    const res = await api.mullvadPing(pingTarget.value, count)
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
    const res = await api.mullvadSpeedtest(mode)
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
                <span class="text-muted-foreground">City:</span>
                {{ status?.city || '—' }}
              </p>
              <p>
                <span class="text-muted-foreground">Server:</span>
                {{ status?.relay || '—' }}
              </p>
            </div>
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
          <CardHeader><CardTitle>Anti-censorship</CardTitle></CardHeader>
          <CardContent class="space-y-3">
            <select v-model="antiMode" class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm">
              <option v-for="m in antiModes" :key="m" :value="m">{{ m }}</option>
            </select>
            <button type="button" class="field-button" :disabled="!!busy" @click="applyAnti">
              {{ busy === 'anti' ? 'Applying…' : 'Apply mode' }}
            </button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader><CardTitle>Tunnel features</CardTitle></CardHeader>
          <CardContent class="space-y-3">
            <label class="flex items-center justify-between gap-3 text-sm">
              <span>Quantum Resistance</span>
              <input v-model="quantumResistant" type="checkbox" class="h-4 w-4" :disabled="!!busy" />
            </label>
            <label class="flex items-center justify-between gap-3 text-sm">
              <span>DAITA</span>
              <input v-model="daita" type="checkbox" class="h-4 w-4" :disabled="!!busy" />
            </label>
            <button type="button" class="field-button" :disabled="!!busy" @click="applyTunnel">
              {{ busy === 'tunnel' ? 'Applying…' : 'Apply tunnel settings' }}
            </button>
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
          <CardTitle>Servers</CardTitle>
          <div class="flex flex-wrap items-center gap-2">
            <div class="flex rounded-md border border-border p-0.5 text-sm">
              <button
                type="button"
                class="rounded px-3 py-1.5"
                :class="serverMode === 'country' ? 'bg-muted font-medium' : 'text-muted-foreground'"
                @click="serverMode = 'country'"
              >
                Country
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
                All servers
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
                  <TableHead>Country</TableHead>
                  <TableHead v-if="serverMode !== 'country'">City</TableHead>
                  <TableHead v-if="serverMode === 'all'">Hostname</TableHead>
                  <TableHead v-if="serverMode === 'all'">IPv4</TableHead>
                  <TableHead>Active</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-if="filteredRelays.length === 0">
                  <TableCell :colspan="serverMode === 'all' ? 6 : serverMode === 'city' ? 4 : 3" class="text-muted-foreground">
                    No servers
                  </TableCell>
                </TableRow>
                <TableRow v-for="r in pageRows" :key="r.hostname + r.cityCode + r.countryCode">
                  <TableCell>
                    <span class="inline-flex items-center gap-2">
                      <CountryFlag :country-code="r.countryCode" :country-name="r.country" />
                      {{ r.country }} ({{ r.countryCode }})
                    </span>
                  </TableCell>
                  <TableCell v-if="serverMode !== 'country'">{{ r.city }} ({{ r.cityCode }})</TableCell>
                  <TableCell v-if="serverMode === 'all'">{{ r.hostname }}</TableCell>
                  <TableCell v-if="serverMode === 'all'">{{ r.ipv4 }}</TableCell>
                  <TableCell>{{ r.active ? 'yes' : '' }}</TableCell>
                  <TableCell>
                    <button
                      type="button"
                      class="field-button"
                      :disabled="!!busy"
                      @click="connectRelay(r)"
                    >
                      {{ busy === `connect:${r.hostname}` ? 'Connecting…' : 'Connect' }}
                    </button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
          <TablePagination
            :page="page"
            :total-pages="totalPages"
            :total-rows="filteredRelays.length"
            :page-size="pageSize"
            @prev="prevPage"
            @next="nextPage"
          />
        </CardContent>
      </Card>
    </div>
  </div>
</template>
