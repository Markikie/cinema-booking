<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div id="app-shell">
    <nav v-if="auth.isAuthenticated" class="top-nav">
      <span>{{ auth.user?.name }} ({{ auth.user?.role }})</span>
      <router-link v-if="auth.isAdmin" to="/admin">Admin</router-link>
      <button @click="logout">Logout</button>
    </nav>
    <router-view />
  </div>
</template>

<style scoped>
.top-nav {
  display: flex;
  gap: 16px;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border, #e5e4e7);
}
.top-nav button {
  margin-left: auto;
}
</style>
