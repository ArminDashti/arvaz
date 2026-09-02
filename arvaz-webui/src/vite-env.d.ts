/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

declare const __APP_VERSION__: string

declare module 'virtual:pwa-register' {
  export function registerSW(options?: {
    immediate?: boolean
    onRegisteredSW?: (swUrl: string, registration: ServiceWorkerRegistration | undefined) => void
  }): (reloadPage?: boolean) => Promise<void>
}
