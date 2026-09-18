import { clearToken, getToken } from '@/lib/auth'

async function authFetch(path: string, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers)
  if (!headers.has('Content-Type') && init?.body && !(init.body instanceof FormData) && !(init.body instanceof Blob) && !(init.body instanceof ArrayBuffer)) {
    headers.set('Content-Type', 'application/json')
  }
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)

  const res = await fetch(path, { ...init, headers })
  if (res.status === 401) {
    clearToken()
    if (!path.includes('/auth/login')) {
      window.location.href = '/login'
    }
    throw new Error('Unauthorized')
  }
  if (!res.ok) {
    if (res.status === 413) {
      throw new Error('upload blocked by proxy/body limit (413)')
    }
    throw new Error(`${res.status} ${res.statusText}`)
  }
  return res
}

/** Count response body bytes without buffering the full payload in one ArrayBuffer. */
async function readBodyByteCount(res: Response, signal?: AbortSignal): Promise<number> {
  if (!res.body) {
    const buf = await res.arrayBuffer()
    return buf.byteLength
  }
  const reader = res.body.getReader()
  let total = 0
  try {
    while (true) {
      if (signal?.aborted) {
        await reader.cancel()
        throw new DOMException('Aborted', 'AbortError')
      }
      const { done, value } = await reader.read()
      if (done) break
      total += value.byteLength
    }
  } catch (e) {
    if (signal?.aborted) {
      throw new DOMException('Aborted', 'AbortError')
    }
    throw e
  }
  return total
}

async function requestJson<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await authFetch(path, init)
  return (await res.json()) as T
}

export const api = {
  login: (username: string, password: string) =>
    requestJson<{ token: string; username: string }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  dockerContainers: () =>
    requestJson<{ containers: DockerContainer[]; error?: string }>('/api/v1/docker/containers'),
  dockerStart: (name: string) =>
    requestJson<{ ok: boolean; error?: string }>(
      `/api/v1/docker/containers/${encodeURIComponent(name)}/start`,
      { method: 'POST' },
    ),
  dockerStop: (name: string) =>
    requestJson<{ ok: boolean; error?: string }>(
      `/api/v1/docker/containers/${encodeURIComponent(name)}/stop`,
      { method: 'POST' },
    ),
  dockerExec: (name: string, command: string) =>
    requestJson<{ ok: boolean; output?: string; error?: string }>(
      `/api/v1/docker/containers/${encodeURIComponent(name)}/exec`,
      { method: 'POST', body: JSON.stringify({ command }) },
    ),
  mullvadStatus: () =>
    requestJson<{ status: MullvadStatus; error?: string }>('/api/v1/mullvad/status'),
  mullvadRelays: () =>
    requestJson<{ relays: MullvadRelay[]; error?: string }>('/api/v1/mullvad/relays'),
  mullvadSetRelay: (country: string, city?: string, hostname?: string) =>
    requestJson<{ ok: boolean; error?: string }>('/api/v1/mullvad/relay', {
      method: 'POST',
      body: JSON.stringify({ country, city, hostname }),
    }),
  mullvadGetAnti: () => requestJson<{ mode: string; error?: string }>('/api/v1/mullvad/anti-censorship'),
  mullvadSetAnti: (mode: string) =>
    requestJson<{ ok: boolean; error?: string }>('/api/v1/mullvad/anti-censorship', {
      method: 'POST',
      body: JSON.stringify({ mode }),
    }),
  mullvadSetTunnel: (opts: { quantumResistant?: boolean; daita?: boolean }) =>
    requestJson<{ ok: boolean; error?: string }>('/api/v1/mullvad/tunnel', {
      method: 'POST',
      body: JSON.stringify(opts),
    }),
  mullvadPing: (target?: string, count?: number) =>
    requestJson<{ result: MullvadPingResult; error?: string }>('/api/v1/mullvad/ping', {
      method: 'POST',
      body: JSON.stringify({ target, count }),
    }),
  mullvadSpeedtest: (mode: 'single' | 'parallel') =>
    requestJson<{ result: MullvadSpeedtestResult; error?: string }>('/api/v1/mullvad/speedtest', {
      method: 'POST',
      body: JSON.stringify({ mode }),
    }),
  windscribeStatus: () =>
    requestJson<{ status: WindscribeStatus; error?: string }>('/api/v1/windscribe/status'),
  windscribeLocations: () =>
    requestJson<{ locations: WindscribeLocation[]; error?: string }>('/api/v1/windscribe/locations'),
  windscribeConnect: (location: string) =>
    requestJson<{ ok: boolean; error?: string }>('/api/v1/windscribe/connect', {
      method: 'POST',
      body: JSON.stringify({ location }),
    }),
  windscribeDisconnect: () =>
    requestJson<{ ok: boolean; error?: string }>('/api/v1/windscribe/disconnect', {
      method: 'POST',
      body: JSON.stringify({}),
    }),
  windscribePing: (target?: string, count?: number) =>
    requestJson<{ result: WindscribePingResult; error?: string }>('/api/v1/windscribe/ping', {
      method: 'POST',
      body: JSON.stringify({ target, count }),
    }),
  windscribeSpeedtest: (mode: 'single' | 'parallel') =>
    requestJson<{ result: WindscribeSpeedtestResult; error?: string }>('/api/v1/windscribe/speedtest', {
      method: 'POST',
      body: JSON.stringify({ mode }),
    }),
  softetherSessions: () =>
    requestJson<{ sessions: SoftEtherSession[]; error?: string }>('/api/v1/softether/sessions'),
  softetherUsers: () =>
    requestJson<{ users: SoftEtherUser[]; error?: string }>('/api/v1/softether/users'),
  softetherUserSessions: (username: string) =>
    requestJson<{ username: string; sessions: SoftEtherSessionLog[]; error?: string }>(
      `/api/v1/softether/users/${encodeURIComponent(username)}/sessions`,
    ),
  softetherIpSessions: (ip: string) =>
    requestJson<{ ip: string; sessions: SoftEtherSessionLog[]; error?: string }>(
      `/api/v1/softether/ip-sessions?ip=${encodeURIComponent(ip)}`,
    ),
  iperfLatency: (signal?: AbortSignal) =>
    requestJson<{ ok: boolean }>('/api/v1/iperf/latency', { signal }),
  iperfDownload: async (bytes: number, signal?: AbortSignal): Promise<number> => {
    const res = await authFetch(`/api/v1/iperf/download?bytes=${bytes}`, { signal })
    return readBodyByteCount(res, signal)
  },
  iperfUpload: async (body: Blob, signal?: AbortSignal): Promise<number> => {
    const res = await authFetch('/api/v1/iperf/upload', {
      method: 'POST',
      headers: { 'Content-Type': 'application/octet-stream' },
      body,
      signal,
    })
    const data = (await res.json()) as { bytesReceived: number }
    return data.bytesReceived
  },
  hostMetrics: () => requestJson<HostMetrics>('/api/v1/system/host-metrics'),
  hostMetricsHistory: (range: HostMetricsHistoryRange) =>
    requestJson<HostMetricsHistoryResponse>(
      `/api/v1/system/host-metrics/history?range=${encodeURIComponent(range)}`,
    ),
}

export type DockerContainer = {
  stackName: string
  containerName: string
  cpuPercent: number
  memoryBytes: number
  memoryGb: number
  diskBytes: number
  diskGb: number
  network: string
  ipPort: string
  haproxyUrl: string
  uptimeSeconds: number
  state: string
}

export type MullvadStatus = {
  raw: string
  connected: boolean
  relay: string
  relayIp?: string
  visibleLocation: string
  publicIp?: string
  country?: string
  city?: string
  features: string
  antiCensorship: string
  quantumResistant?: boolean
  daita?: boolean
}

export type MullvadRelay = {
  country: string
  countryCode: string
  city: string
  cityCode: string
  hostname: string
  ipv4: string
  active: boolean
}

export type MullvadPingResult = {
  target: string
  count: number
  packetLossPercent: number
  avgMs: number
}

export type MullvadSpeedtestResult = {
  mode: string
  raw: string
  downloadMbps?: number
  uploadMbps?: number
  latencyMs?: number
  parsedOk?: boolean
}

export type WindscribeStatus = {
  raw: string
  connected: boolean
  location?: string
  publicIp?: string
  country?: string
  city?: string
  protocol?: string
}

export type WindscribeLocation = {
  region: string
  city: string
  nickname: string
  label: string
  countryCode?: string
  active: boolean
}

export type WindscribePingResult = {
  target: string
  count: number
  packetLossPercent: number
  avgMs: number
}

export type WindscribeSpeedtestResult = {
  mode: string
  raw: string
  downloadMbps?: number
  uploadMbps?: number
  latencyMs?: number
  parsedOk?: boolean
}

export type SoftEtherSession = {
  username: string
  clientIp: string
  lastIsp?: string
  ispLogo?: string
  sessionName?: string
  downloadBytes: number
  uploadBytes: number
  transferBytes?: number
  downloadMbps?: number | null
  uploadMbps?: number | null
  sessionDurationSeconds?: number
  connectedAt?: string
  sessionKey?: string
}

export type SoftEtherUser = {
  username: string
  lastLogin: string
  downloadBytes: number
  uploadBytes: number
  lastIp?: string
  lastIsp?: string
  ispLogo?: string
  trafficYesterdayBytes?: number
  trafficWeekBytes?: number
  trafficMonthBytes?: number
  trafficTotalBytes?: number
  trafficYesterdaySeconds?: number
  trafficWeekSeconds?: number
  trafficMonthSeconds?: number
}

export type SoftEtherSessionLog = {
  username?: string
  downloadBytes: number
  uploadBytes: number
  ip: string
  isp?: string
  ispLogo?: string
  durationSeconds: number
  connectedAt: string
  disconnectedAt?: string
}

export type HostMetrics = {
  timestamp?: string
  cpuCores?: number[]
  memory?: {
    totalGb: number
    usedGb: number
    availableGb: number
  }
  disk?: {
    totalGb: number
    usedGb: number
    freeGb: number
  }
  network?: {
    downloadMbps: number
    uploadMbps: number
  }
  error?: string
}

export type HostMetricsHistoryRange =
  | 'today'
  | 'yesterday'
  | 'this_week'
  | 'last_week'
  | 'this_month'
  | 'all'

export type HostMetricsHistoryPoint = {
  ts: string
  cpuPct: number
  memUsedGb: number
  memTotalGb: number
  diskUsedGb: number
  diskTotalGb: number
  netDownMbps: number
  netUpMbps: number
}

export type HostMetricsHistoryResponse = {
  range?: string
  points?: HostMetricsHistoryPoint[]
  error?: string
}
