<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Server } from 'lucide-vue-next'
import { api, type HostMetrics } from '@/api/client'
import { avgCpu, barWidth, formatMbps, formatPct, usagePct } from '@/lib/hostMetricsFormat'
import { MANAGED_HOSTS } from '@/lib/hosts'

const REFRESH_MS = 4_000

const metricsByHost = ref<Record<string, HostMetrics | null>>({})
const errorByHost = ref<Record<string, string>>({})
const loading = ref(false)
let timer: ReturnType<typeof setInterval> | undefined

function hostMetrics(hostId: string): HostMetrics | null {
  return metricsByHost.value[hostId] ?? null
}

function hostError(hostId: string): string {
  return errorByHost.value[hostId] ?? ''
}

function cpuPct(hostId: string): number {
  return avgCpu(hostMetrics(hostId)?.cpuCores)
}

function memoryPct(hostId: string): number {
  const m = hostMetrics(hostId)?.memory
  return usagePct(m?.usedGb, m?.totalGb)
}

function diskPct(hostId: string): number {
  const d = hostMetrics(hostId)?.disk
  return usagePct(d?.usedGb, d?.totalGb)
}

async function refreshMetrics() {
  if (loading.value) return
  loading.value = true
  try {
    for (const host of MANAGED_HOSTS) {
      if (!host.monitorHealth) continue
      try {
        const res = await api.hostMetrics()
        if (res.error) {
          errorByHost.value[host.id] = res.error
          metricsByHost.value[host.id] = null
        } else {
          errorByHost.value[host.id] = ''
          metricsByHost.value[host.id] = res
        }
      } catch (e) {
        errorByHost.value[host.id] = e instanceof Error ? e.message : 'Failed to load'
        metricsByHost.value[host.id] = null
      }
    }
  } finally {
    loading.value = false
  }
}

const hasAnyMetrics = computed(() =>
  MANAGED_HOSTS.some((h) => hostMetrics(h.id) && !hostError(h.id)),
)

onMounted(() => {
  void refreshMetrics()
  timer = setInterval(() => void refreshMetrics(), REFRESH_MS)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <section class="site-sidebar-hosts" aria-label="Managed hosts">
    <div class="site-sidebar-hosts-head">
      <Server class="site-sidebar-hosts-icon" :size="14" aria-hidden="true" />
      <span>Hosts</span>
      <span v-if="loading && hasAnyMetrics" class="site-sidebar-hosts-live">live</span>
    </div>
    <ul class="site-sidebar-hosts-list">
      <li
        v-for="host in MANAGED_HOSTS"
        :key="host.id"
        class="site-sidebar-host-card"
        :style="{
          '--host-accent': host.accent,
          '--host-accent-soft': host.accentSoft,
        }"
      >
        <div class="site-sidebar-host-header">
          <span
            class="site-sidebar-host-dot"
            :class="{
              'is-online': hostMetrics(host.id) && !hostError(host.id),
              'is-offline': hostError(host.id),
              'is-checking': loading && !hostMetrics(host.id) && !hostError(host.id),
            }"
          />
          <div class="site-sidebar-host-text">
            <span class="site-sidebar-host-name">{{ host.name }}</span>
            <span class="site-sidebar-host-sub">{{ host.subtitle }}</span>
          </div>
        </div>

        <p v-if="hostError(host.id)" class="site-sidebar-host-error">{{ hostError(host.id) }}</p>

        <dl v-else class="site-sidebar-host-metrics">
          <div class="site-sidebar-metric site-sidebar-metric-cpu">
            <dt>CPU</dt>
            <dd>
              <div class="site-sidebar-metric-bar" aria-hidden="true">
                <span :style="{ width: barWidth(cpuPct(host.id)) }" />
              </div>
              <span class="site-sidebar-metric-val">{{ formatPct(cpuPct(host.id)) }}</span>
            </dd>
          </div>
          <div class="site-sidebar-metric site-sidebar-metric-mem">
            <dt>Mem</dt>
            <dd>
              <div class="site-sidebar-metric-bar" aria-hidden="true">
                <span :style="{ width: barWidth(memoryPct(host.id)) }" />
              </div>
              <span class="site-sidebar-metric-val">{{ formatPct(memoryPct(host.id)) }}</span>
            </dd>
          </div>
          <div class="site-sidebar-metric site-sidebar-metric-disk">
            <dt>Disk</dt>
            <dd>
              <div class="site-sidebar-metric-bar" aria-hidden="true">
                <span :style="{ width: barWidth(diskPct(host.id)) }" />
              </div>
              <span class="site-sidebar-metric-val">{{ formatPct(diskPct(host.id)) }}</span>
            </dd>
          </div>
          <div class="site-sidebar-metric site-sidebar-metric-net">
            <dt>Net</dt>
            <dd class="site-sidebar-metric-net-values">
              <span class="site-sidebar-net-down" title="Download">
                ↓ {{ formatMbps(hostMetrics(host.id)?.network?.downloadMbps) }}
              </span>
              <span class="site-sidebar-net-up" title="Upload">
                ↑ {{ formatMbps(hostMetrics(host.id)?.network?.uploadMbps) }}
              </span>
            </dd>
          </div>
        </dl>
      </li>
    </ul>
  </section>
</template>
