import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/',
    name: 'ShowList',
    component: () => import('../views/ShowListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/shows/:showtimeId',
    name: 'SeatMap',
    component: () => import('../views/SeatMapView.vue'),
    props: true,
    meta: { requiresAuth: true },
  },
  {
    path: '/booking/:bookingId',
    name: 'ConfirmBooking',
    component: () => import('../views/ConfirmBookingView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('../views/AdminDashboardView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const session = JSON.parse(localStorage.getItem('cinema-booking-session') || '{}')
  const isAuthenticated = Boolean(session.token)
  const isAdmin = session.user?.role === 'ADMIN'

  if (to.meta.requiresAuth && !isAuthenticated) return '/login'
  if (to.meta.requiresAdmin && !isAdmin) return '/'
  if (to.meta.guestOnly && isAuthenticated) return '/'
})

export default router
