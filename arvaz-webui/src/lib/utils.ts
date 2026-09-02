import type { ClassValue } from 'clsx'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Placeholder for missing IP / ISP cells in SoftEther tables. */
export const emptyGridValue = 'error'

export function formatBytes(n: number | undefined | null): string {
  if (n == null || Number.isNaN(n)) return '—'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let v = n
  let i = 0
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`
}

export function formatBps(n: number | undefined | null): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${formatBytes(n)}/s`
}

export function formatDuration(seconds: number | undefined | null): string {
  if (seconds == null || Number.isNaN(seconds)) return '—'
  const s = Math.max(0, Math.floor(seconds))
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

/**
 * Prefer stored duration; if missing or zero, derive from connected/disconnected
 * timestamps (still-open sessions use now).
 */
export function resolveDurationSeconds(opts: {
  durationSeconds?: number | null
  connectedAt?: string | Date | null
  disconnectedAt?: string | Date | null
  now?: Date
}): number {
  const stored = opts.durationSeconds
  if (stored != null && !Number.isNaN(stored) && stored > 0) {
    return Math.floor(stored)
  }
  const start = typeof opts.connectedAt === 'string' || opts.connectedAt == null
    ? parseFlexibleDate(opts.connectedAt)
    : opts.connectedAt
  if (!start) return stored != null && !Number.isNaN(stored) ? Math.max(0, Math.floor(stored)) : 0
  const end =
    typeof opts.disconnectedAt === 'string' || opts.disconnectedAt == null
      ? parseFlexibleDate(opts.disconnectedAt) ?? opts.now ?? new Date()
      : opts.disconnectedAt
  const seconds = Math.floor((end.getTime() - start.getTime()) / 1000)
  return Math.max(0, seconds)
}

export function formatPercent(n: number | undefined | null): string {
  if (n == null || Number.isNaN(n)) return '—'
  return `${n.toFixed(1)}%`
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return '—'
  const d = parseFlexibleDate(value)
  if (!d) return '—'
  return d.toLocaleString()
}

/** HH:MM only (local), no seconds. */
export function formatClockHHMM(value: string | Date | null | undefined): string {
  const d = typeof value === 'string' || value == null ? parseFlexibleDate(value) : value
  if (!d) return '—'
  return `${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}

/**
 * Last login: today → HH:MM; yesterday → Yesterday HH:MM; older → YYYY-MM-DD HH:MM.
 * No seconds.
 */
export function formatLastLogin(value: string | null | undefined): string {
  if (!value) return '—'
  const d = parseFlexibleDate(value)
  if (!d) return value.includes(':') ? stripSeconds(value) : value
  const clock = `${pad2(d.getHours())}:${pad2(d.getMinutes())}`
  const today = startOfLocalDay(new Date())
  const day = startOfLocalDay(d)
  const diffDays = Math.round((today.getTime() - day.getTime()) / 86400000)
  if (diffDays === 0) return clock
  if (diffDays === 1) return `Yesterday ${clock}`
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())} ${clock}`
}

export function softetherIpLogsPath(ip: string): string {
  return `/softether/ips/${encodeURIComponent(ip)}`
}

export type SortDir = 'asc' | 'desc'

export function compareSortValues(a: unknown, b: unknown, dir: SortDir): number {
  const mul = dir === 'asc' ? 1 : -1
  if (a == null && b == null) return 0
  if (a == null) return 1 * mul
  if (b == null) return -1 * mul
  if (typeof a === 'number' && typeof b === 'number') {
    return (a - b) * mul
  }
  const as = String(a)
  const bs = String(b)
  const an = Number(as)
  const bn = Number(bs)
  if (as !== '' && bs !== '' && !Number.isNaN(an) && !Number.isNaN(bn) && /^-?\d+(\.\d+)?$/.test(as) && /^-?\d+(\.\d+)?$/.test(bs)) {
    return (an - bn) * mul
  }
  const ad = Date.parse(as)
  const bd = Date.parse(bs)
  if (!Number.isNaN(ad) && !Number.isNaN(bd) && looksLikeDate(as) && looksLikeDate(bs)) {
    return (ad - bd) * mul
  }
  return as.localeCompare(bs, undefined, { numeric: true, sensitivity: 'base' }) * mul
}

function looksLikeDate(s: string): boolean {
  return /\d{4}-\d{2}-\d{2}|\d{1,2}:\d{2}/.test(s) || !Number.isNaN(Date.parse(s))
}

function pad2(n: number): string {
  return String(n).padStart(2, '0')
}

function startOfLocalDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

function stripSeconds(raw: string): string {
  return raw.replace(/(\d{1,2}:\d{2}):\d{2}/, '$1')
}

export function parseFlexibleDate(value: string | null | undefined): Date | null {
  if (!value) return null
  const trimmed = value.trim()
  if (!trimmed || trimmed === '-' || /^none$/i.test(trimmed)) return null
  const iso = Date.parse(trimmed)
  if (!Number.isNaN(iso)) return new Date(iso)
  // SoftEther-ish: 2026-08-12 15:54:33 or 2026/08/12 15:54:33
  const m = trimmed.match(
    /^(\d{4})[\/\-.](\d{1,2})[\/\-.](\d{1,2})(?:[ T](\d{1,2}):(\d{2})(?::(\d{2}))?)?/,
  )
  if (m) {
    const y = Number(m[1])
    const mo = Number(m[2]) - 1
    const day = Number(m[3])
    const h = Number(m[4] || 0)
    const mi = Number(m[5] || 0)
    const s = Number(m[6] || 0)
    const d = new Date(y, mo, day, h, mi, s)
    if (!Number.isNaN(d.getTime())) return d
  }
  return null
}
