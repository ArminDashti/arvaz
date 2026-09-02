import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '@/lib/auth'
import AppShell from '@/components/AppShell.vue'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'
import ContainersView from '@/views/ContainersView.vue'
import MullvadView from '@/views/MullvadView.vue'
import IperfView from '@/views/IperfView.vue'
import SoftEtherOnlineView from '@/views/SoftEtherOnlineView.vue'
import SoftEtherUsersView from '@/views/SoftEtherUsersView.vue'
import SoftEtherUserLogsView from '@/views/SoftEtherUserLogsView.vue'
import SoftEtherIpLogsView from '@/views/SoftEtherIpLogsView.vue'
import AboutMeView from '@/views/AboutMeView.vue'
import HostMonitorView from '@/views/HostMonitorView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    {
      path: '/',
      component: AppShell,
      children: [
        { path: '', name: 'dashboard', component: DashboardView },
        { path: 'containers', name: 'containers', component: ContainersView },
        { path: 'host', name: 'host', component: HostMonitorView },
        { path: 'mullvad', name: 'mullvad', component: MullvadView },
        { path: 'iperf', name: 'iperf', component: IperfView },
        { path: 'softether/users/online', name: 'softether-online', component: SoftEtherOnlineView },
        { path: 'softether/users/:username/logs', name: 'softether-user-logs', component: SoftEtherUserLogsView },
        { path: 'softether/users', name: 'softether-users', component: SoftEtherUsersView },
        { path: 'softether/ips/:ip', name: 'softether-ip-logs', component: SoftEtherIpLogsView },
        { path: 'about-me', name: 'about-me', component: AboutMeView },
      ],
    },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) {
    if (to.path === '/login' && isAuthenticated()) return '/'
    return true
  }
  if (!isAuthenticated()) return '/login'
  return true
})

export default router
