<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

const router = useRouter()
const auth = useAuthStore()
const error = ref('')
const buttonEl = ref(null)

function waitForGoogleScript() {
  return new Promise((resolve, reject) => {
    const start = Date.now()
    const check = () => {
      if (window.google?.accounts?.id) return resolve()
      if (Date.now() - start > 10000) return reject(new Error('Google script failed to load'))
      setTimeout(check, 100)
    }
    check()
  })
}

async function handleCredentialResponse(response) {
  try {
    const { data } = await api.post('/api/auth/login', { id_token: response.credential })
    auth.setSession(data.token, data.user)
    router.push(auth.isAdmin ? '/admin' : '/')
  } catch (err) {
    error.value = 'Login failed. Please try again.'
    console.error(err)
  }
}

onMounted(async () => {
  try {
    await waitForGoogleScript()
    window.google.accounts.id.initialize({
      client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
      callback: handleCredentialResponse,
    })
    window.google.accounts.id.renderButton(buttonEl.value, {
      theme: 'outline',
      size: 'large',
    })
  } catch (err) {
    error.value = 'Could not load Google Sign-In.'
    console.error(err)
  }
})
</script>

<template>
  <div class="login-page">
    <h1>Cinema Booking</h1>
    <p>Sign in with Google to select seats and book tickets.</p>
    <div ref="buttonEl"></div>
    <p v-if="error" class="error">{{ error }}</p>
  </div>
</template>

<style scoped>
.login-page {
  max-width: 360px;
  margin: 120px auto;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}
.error {
  color: #d33;
}
</style>
