import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      component: () => import('@/components/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
        { path: 'meters', name: 'meters', component: () => import('@/views/MetersView.vue') },
        {
          path: 'meters/:meterId',
          name: 'meter-detail',
          component: () => import('@/views/MeterDetailView.vue'),
        },
        {
          path: 'anomalies',
          name: 'anomalies',
          component: () => import('@/views/AnomaliesView.vue'),
        },
        {
          path: 'anomalies/:id',
          name: 'investigation',
          component: () => import('@/views/InvestigationView.vue'),
        },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

// Everything except /login needs a session; a logged-in user has no business on /login.
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) return { name: 'login' }
  if (to.meta.guest && auth.isAuthenticated) return { name: 'dashboard' }
})

export default router
