# Arvaz WebUI

Vue 3 + Tailwind + shadcn-vue host dashboard for Arvaz API.

## Stack

- Vue 3, TypeScript, Vite
- Tailwind CSS v4, shadcn-vue, Inter font
- PWA (vite-plugin-pwa)
- Dark theme
- Icon sidebar (Dashboard, Docker, VPN, SE Online, SE Users, About Me)

## Local dev

```bash
npm install
npm run dev
```

Dev server proxies `/api` to `http://127.0.0.1:8090`.

Login: `armin` / `dopadopa1234`

## Build

```bash
npm run build
```

Serve `dist/` behind HAProxy/nginx on T3.
