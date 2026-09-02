export function formatPct(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n.toFixed(0)}%`
}

export function formatMbps(n: number | undefined): string {
  if (n == null || Number.isNaN(n)) return '—'
  return n >= 10 ? `${n.toFixed(1)}` : `${n.toFixed(2)}`
}

export function usagePct(used: number | undefined, total: number | undefined): number {
  if (!total || total <= 0 || used == null) return 0
  return Math.max(0, Math.min(100, (used / total) * 100))
}

export function avgCpu(cores: number[] | undefined): number {
  if (!cores?.length) return 0
  return cores.reduce((sum, v) => sum + v, 0) / cores.length
}

export function barWidth(pct: number | undefined): string {
  return `${Math.max(0, Math.min(100, pct ?? 0))}%`
}
