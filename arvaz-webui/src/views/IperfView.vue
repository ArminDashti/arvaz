<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api/client'
import {
  IPERF_DEFAULT_BYTES,
  IPERF_SIZE_CHOICES,
  IPERF_STREAM_CHOICES,
  measureIperf,
  type IperfDirection,
  type IperfPhase,
  type IperfResult,
} from '@/lib/iperf'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const selectClass = 'rounded-md border border-border bg-background px-2 py-1.5 text-sm text-foreground'

const bytesPerStream = ref(IPERF_DEFAULT_BYTES)
const streams = ref(1)
const direction = ref<IperfDirection>('both')
const busy = ref(false)
const phase = ref<IperfPhase | ''>('')
const error = ref('')
const result = ref<IperfResult | null>(null)
let abortController: AbortController | null = null

async function run() {
  abortController?.abort()
  abortController = new AbortController()
  busy.value = true
  phase.value = ''
  error.value = ''
  result.value = null
  try {
    result.value = await measureIperf({
      bytesPerStream: bytesPerStream.value,
      streams: streams.value,
      direction: direction.value,
      signal: abortController.signal,
      onPhase: (p) => {
        phase.value = p
      },
      latency: (signal) => api.iperfLatency(signal),
      download: (bytes, signal) => api.iperfDownload(bytes, signal),
      upload: (body, signal) => api.iperfUpload(body, signal),
    })
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') {
      error.value = 'Cancelled'
    } else {
      error.value = e instanceof Error ? e.message : 'Iperf test failed'
    }
  } finally {
    busy.value = false
    phase.value = ''
    abortController = null
  }
}

function cancel() {
  abortController?.abort()
}

function formatMbps(n: number | null | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n >= 10 ? n.toFixed(0) : n.toFixed(1)} Mbps`
}

function sizeLabel(bytes: number): string {
  return `${bytes / (1024 * 1024)} MiB`
}

function phaseLabel(p: IperfPhase | ''): string {
  if (p === 'latency') return 'Measuring latency…'
  if (p === 'download') return 'Downloading…'
  if (p === 'upload') return 'Uploading…'
  return 'Running…'
}
</script>

<template>
  <div class="space-y-4">
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card>
      <CardHeader class="space-y-3">
        <div>
          <CardTitle>LAN Iperf (HTTP)</CardTitle>
          <p class="mt-1 text-sm text-muted-foreground">
            Measures browser ↔ API speed on your LAN/VPN. No internet. Choose size, streams, and
            direction, then Run.
          </p>
        </div>
        <div class="flex flex-wrap items-end gap-3">
          <label class="flex flex-col gap-1 text-sm text-muted-foreground">
            Size
            <select v-model.number="bytesPerStream" :class="selectClass" :disabled="busy">
              <option v-for="n in IPERF_SIZE_CHOICES" :key="n" :value="n">{{ sizeLabel(n) }}</option>
            </select>
          </label>
          <label class="flex flex-col gap-1 text-sm text-muted-foreground">
            Streams
            <select v-model.number="streams" :class="selectClass" :disabled="busy">
              <option v-for="n in IPERF_STREAM_CHOICES" :key="n" :value="n">{{ n }}</option>
            </select>
          </label>
          <label class="flex flex-col gap-1 text-sm text-muted-foreground">
            Direction
            <select v-model="direction" :class="selectClass" :disabled="busy">
              <option value="both">Both</option>
              <option value="download">Download only</option>
              <option value="upload">Upload only</option>
            </select>
          </label>
          <div class="flex flex-wrap gap-2">
            <button type="button" class="field-button" :disabled="busy" @click="run">
              {{ busy ? phaseLabel(phase) : 'Run' }}
            </button>
            <button type="button" class="field-button" :disabled="!busy" @click="cancel">
              Cancel
            </button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div v-if="result" class="grid gap-4 sm:grid-cols-3">
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Download</p>
            <p class="mt-1 text-3xl font-semibold text-emerald-500">
              {{ formatMbps(result.downloadMbps) }}
            </p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Upload</p>
            <p class="mt-1 text-3xl font-semibold text-red-500">
              {{ formatMbps(result.uploadMbps) }}
            </p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Latency</p>
            <p class="mt-1 text-3xl font-semibold">
              {{ result.latencyMs != null ? `${result.latencyMs.toFixed(0)} ms` : '—' }}
            </p>
          </div>
          <p class="sm:col-span-3 text-xs text-muted-foreground">
            Direction: {{ result.direction }} · Streams: {{ result.streams }} ·
            {{ (result.bytesPerStream / (1024 * 1024)).toFixed(0) }} MiB per stream
          </p>
        </div>
        <p v-else class="text-sm text-muted-foreground">
          {{ busy ? phaseLabel(phase) : 'Configure options and Run to see Download, Upload, and Latency.' }}
        </p>
      </CardContent>
    </Card>
  </div>
</template>
