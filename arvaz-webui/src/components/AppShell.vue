<script setup lang="ts">
import { Activity, Boxes, Gauge, LayoutDashboard, Shield, User, Users, Wifi } from 'lucide-vue-next'
import { RouterLink, RouterView, useRouter } from 'vue-router'
import SidebarHosts from '@/components/SidebarHosts.vue'
import { clearToken } from '@/lib/auth'

const router = useRouter()
const appVersion = __APP_VERSION__

function logout() {
  clearToken()
  router.push('/login')
}

const links = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard },
  { to: '/containers', label: 'Docker', icon: Boxes },
  { to: '/host', label: 'Host', icon: Activity },
  { to: '/mullvad', label: 'Mullvad', icon: Shield },
  { to: '/iperf', label: 'Iperf', icon: Gauge },
  { to: '/softether/users/online', label: 'SE Online', icon: Wifi },
  { to: '/softether/users', label: 'SE Users', icon: Users },
  { to: '/about-me', label: 'About Me', icon: User },
]
</script>

<template>
  <div class="app-shell">
    <aside class="site-sidebar">
      <RouterLink to="/" class="site-logo">Arvaz</RouterLink>
      <nav class="site-nav" aria-label="Primary">
        <RouterLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="site-nav-link"
          active-class=""
          exact-active-class="is-active"
        >
          <component :is="link.icon" class="site-nav-icon" :size="18" aria-hidden="true" />
          <span>{{ link.label }}</span>
        </RouterLink>
      </nav>
      <SidebarHosts />
      <div class="site-sidebar-logout-wrap">
        <button type="button" class="field-button site-sidebar-logout" @click="logout">Logout</button>
        <p class="site-sidebar-company">Dashti Technologies LLC</p>
        <p class="site-sidebar-version">v{{ appVersion }}</p>
      </div>
    </aside>

    <div class="site-content">
      <main class="site-main">
        <RouterView />
      </main>
    </div>
  </div>
</template>
