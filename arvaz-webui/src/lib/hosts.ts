export type ManagedHost = {
  id: string
  name: string
  subtitle: string
  accent: string
  accentSoft: string
  /** When true, health is checked via the current API host-metrics endpoint. */
  monitorHealth?: boolean
}

export const MANAGED_HOSTS: ManagedHost[] = [
  {
    id: 't3',
    name: 'Irancell-T3',
    subtitle: '2.144.27.124',
    accent: 'oklch(0.78 0.16 195)',
    accentSoft: 'oklch(0.32 0.06 195)',
    monitorHealth: true,
  },
]
