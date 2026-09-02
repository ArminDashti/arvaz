export function flagEmoji(countryCode: string | undefined): string {
  if (!countryCode || countryCode.length !== 2) return '🏳️'
  const cc = countryCode.toUpperCase()
  return String.fromCodePoint(...[...cc].map((c) => 127397 + c.charCodeAt(0)))
}

export function countryCodeFromRelay(hostname: string | undefined): string | undefined {
  if (!hostname) return undefined
  const prefix = hostname.split('-')[0]
  if (prefix && /^[a-z]{2}$/i.test(prefix)) return prefix.toLowerCase()
  return undefined
}
