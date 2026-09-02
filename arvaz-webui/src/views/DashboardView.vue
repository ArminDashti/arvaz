<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, type DockerContainer, type MullvadStatus, type SoftEtherSession, type SoftEtherUser } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { countryCodeFromRelay, flagEmoji } from '@/lib/flag'

const error = ref('')
const loading = ref(false)
const status = ref<MullvadStatus | null>(null)
const sessions = ref<SoftEtherSession[]>([])
const users = ref<SoftEtherUser[]>([])
const containers = ref<DockerContainer[]>([])

const mullvadFlag = computed(() => flagEmoji(countryCodeFromRelay(status.value?.relay)))

const onlineCount = computed(() => sessions.value.length)
const userCount = computed(() => users.value.length)
const runningCount = computed(() => containers.value.filter((c) => c.state === 'running').length)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [st, se, us, dk] = await Promise.all([
      api.mullvadStatus(),
      api.softetherSessions(),
      api.softetherUsers(),
      api.dockerContainers(),
    ])
    status.value = st.status ?? null
    sessions.value = se.sessions ?? []
    users.value = us.users ?? []
    containers.value = dk.containers ?? []
    error.value = st.error || se.error || us.error || dk.error || ''
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load dashboard'
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-semibold tracking-tight">Dashboard</h1>
      <button type="button" class="field-button" :disabled="loading" @click="load">
        {{ loading ? 'Refreshing…' : 'Refresh' }}
      </button>
    </div>
    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <div class="grid gap-4 lg:grid-cols-3">
      <Card>
        <CardHeader><CardTitle>Mullvad</CardTitle></CardHeader>
        <CardContent class="space-y-3">
          <p
            class="text-2xl font-semibold"
            :class="status?.connected ? 'text-emerald-500' : 'text-red-500'"
          >
            {{ status?.connected ? 'Connected' : 'Disconnected' }}
          </p>
          <p class="text-4xl leading-none" :title="status?.country || ''">{{ mullvadFlag }}</p>
          <div class="space-y-1 text-sm">
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
        <CardHeader><CardTitle>SoftEther</CardTitle></CardHeader>
        <CardContent class="space-y-1">
          <p class="text-3xl font-semibold">{{ onlineCount }}</p>
          <p class="text-sm text-muted-foreground">online of {{ userCount }} users</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle>Docker</CardTitle></CardHeader>
        <CardContent class="space-y-1">
          <p class="text-3xl font-semibold">{{ runningCount }}</p>
          <p class="text-sm text-muted-foreground">running of {{ containers.length }}</p>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
