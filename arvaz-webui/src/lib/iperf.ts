export const IPERF_DEFAULT_BYTES = 8 * 1024 * 1024
export const IPERF_SIZE_CHOICES = [
  1 * 1024 * 1024,
  8 * 1024 * 1024,
  32 * 1024 * 1024,
  64 * 1024 * 1024,
] as const
export const IPERF_STREAM_CHOICES = [1, 2, 4, 8] as const

export type IperfDirection = 'both' | 'download' | 'upload'
export type IperfPhase = 'latency' | 'download' | 'upload'

export type IperfResult = {
  streams: number
  direction: IperfDirection
  downloadMbps: number | null
  uploadMbps: number | null
  latencyMs: number | null
  bytesPerStream: number
}

function mbps(totalBytes: number, elapsedMs: number): number {
  if (elapsedMs <= 0) return 0
  return (totalBytes * 8) / (elapsedMs / 1000) / 1_000_000
}

function throwIfAborted(signal?: AbortSignal) {
  if (signal?.aborted) {
    throw new DOMException('Aborted', 'AbortError')
  }
}

/** Build one zero-filled Blob of the given size (reused across parallel uploads). */
export function buildZeroBlob(bytes: number): Blob {
  return new Blob([new Uint8Array(bytes)])
}

export async function measureIperf(opts: {
  bytesPerStream?: number
  streams?: number
  direction?: IperfDirection
  signal?: AbortSignal
  onPhase?: (phase: IperfPhase) => void
  latency: (signal?: AbortSignal) => Promise<unknown>
  download: (bytes: number, signal?: AbortSignal) => Promise<number>
  upload: (body: Blob, signal?: AbortSignal) => Promise<number>
}): Promise<IperfResult> {
  const bytesPerStream = opts.bytesPerStream ?? IPERF_DEFAULT_BYTES
  const streams = opts.streams ?? 1
  const direction = opts.direction ?? 'both'

  throwIfAborted(opts.signal)
  opts.onPhase?.('latency')
  let latencyMs: number | null = null
  try {
    const t0 = performance.now()
    await opts.latency(opts.signal)
    latencyMs = performance.now() - t0
  } catch (e) {
    if (opts.signal?.aborted || (e instanceof DOMException && e.name === 'AbortError')) {
      throw e
    }
    latencyMs = null
  }

  let downloadMbps: number | null = null
  if (direction === 'both' || direction === 'download') {
    throwIfAborted(opts.signal)
    opts.onPhase?.('download')
    const downloadStart = performance.now()
    const downloadBytes = (
      await Promise.all(
        Array.from({ length: streams }, () => opts.download(bytesPerStream, opts.signal)),
      )
    ).reduce((a, b) => a + b, 0)
    downloadMbps = mbps(downloadBytes, performance.now() - downloadStart)
  }

  let uploadMbps: number | null = null
  if (direction === 'both' || direction === 'upload') {
    throwIfAborted(opts.signal)
    opts.onPhase?.('upload')
    const payload = buildZeroBlob(bytesPerStream)
    const uploadStart = performance.now()
    const uploadBytes = (
      await Promise.all(Array.from({ length: streams }, () => opts.upload(payload, opts.signal)))
    ).reduce((a, b) => a + b, 0)
    uploadMbps = mbps(uploadBytes, performance.now() - uploadStart)
  }

  return {
    streams,
    direction,
    downloadMbps,
    uploadMbps,
    latencyMs,
    bytesPerStream,
  }
}
