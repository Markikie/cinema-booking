<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'

const route = useRoute()
const router = useRouter()

const bookingId = route.params.bookingId
const expiresAt = Number(route.query.expiresAt) || null
const now = ref(Date.now())
const status = ref('pending') // pending | paying | success | expired | error
const error = ref('')
let tickInterval = null

const secondsLeft = computed(() => {
  if (!expiresAt) return null
  return Math.max(0, Math.floor((expiresAt - now.value) / 1000))
})

async function payNow() {
  status.value = 'paying'
  try {
    await api.post('/api/bookings/confirm-payment', { booking_id: bookingId })
    status.value = 'success'
  } catch (err) {
    if (err.response?.status === 410) {
      status.value = 'expired'
    } else {
      status.value = 'error'
      error.value = err.response?.data?.error || 'Payment failed'
    }
  }
}

onMounted(() => {
  tickInterval = setInterval(() => {
    now.value = Date.now()
    if (secondsLeft.value === 0 && status.value === 'pending') {
      status.value = 'expired'
    }
  }, 1000)
})

onBeforeUnmount(() => clearInterval(tickInterval))
</script>

<template>
  <div class="confirm-page">
    <h1>Confirm Booking</h1>
    <p>Booking ID: {{ bookingId }}</p>

    <template v-if="status === 'pending' || status === 'paying'">
      <p v-if="secondsLeft !== null">Time left to pay: {{ secondsLeft }}s</p>
      <button :disabled="status === 'paying'" @click="payNow">
        {{ status === 'paying' ? 'Processing…' : 'Pay now (mock)' }}
      </button>
    </template>

    <template v-else-if="status === 'success'">
      <p class="success">Booking confirmed! Enjoy the movie.</p>
      <button @click="router.push('/')">Back to home</button>
    </template>

    <template v-else-if="status === 'expired'">
      <p class="error">Your seats were released because time ran out.</p>
      <button @click="router.push('/')">Choose seats again</button>
    </template>

    <template v-else-if="status === 'error'">
      <p class="error">{{ error }}</p>
      <button @click="payNow">Try again</button>
    </template>
  </div>
</template>

<style scoped>
.confirm-page {
  max-width: 420px;
  margin: 80px auto;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}
.success {
  color: #2e7d32;
}
.error {
  color: #d33;
}
</style>
