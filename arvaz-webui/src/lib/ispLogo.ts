const RULES: { key: string; needles: string[] }[] = [
  { key: 'zitel', needles: ['zitel', 'pasargad arian', "tose'h fanavari", 'toseh fanavari'] },
  { key: 'irancell', needles: ['irancell', 'iran cell'] },
  { key: 'mobin-net', needles: ['mobin net', 'mobinnet'] },
  { key: 'mci', needles: ['mobile communication company of iran', 'mobile telecommunication', 'hamrah-e', 'hamrahe aval'] },
  { key: 'respina', needles: ['respina'] },
  { key: 'shatel', needles: ['shatel', 'aria shatel'] },
]

export function ispLogoKey(name?: string, fromApi?: string): string | undefined {
  const api = fromApi?.trim()
  if (api) return api
  const n = (name || '').trim().toLowerCase().replace(/\u2019/g, "'")
  if (!n) return undefined
  if (n === 'mci' || n.startsWith('mci ')) return 'mci'
  for (const rule of RULES) {
    if (rule.needles.some((needle) => n.includes(needle))) return rule.key
  }
  return undefined
}

export function ispLogoSrc(name?: string, fromApi?: string): string | undefined {
  const key = ispLogoKey(name, fromApi)
  return key ? `/isp-logos/${key}.png?v=3` : undefined
}
