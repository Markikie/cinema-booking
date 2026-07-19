<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSeatMapStore } from '@/stores/seatMap'
import { SeatSocket } from '@/services/socket'
import api from '@/services/api'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const seatMap = useSeatMapStore()

const showtimeId = route.params.showtimeId
const showtime = ref(null)
const error = ref('')
const now = ref(Date.now())
let socket = null
let tickInterval = null

const rows = computed(() => {
  const grouped = {}
  for (const seat of seatMap.seats) {
    grouped[seat.row] ??= []
    grouped[seat.row].push(seat)
  }
  return Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b))
})

const secondsLeft = computed(() => {
  if (!seatMap.lockExpiresAt) return null
  return Math.max(0, Math.floor((seatMap.lockExpiresAt - now.value) / 1000))
})

const pageTitle = computed(() => showtime.value?.movie_name || 'Select seats')
const showtimeDetail = computed(() => {
  if (!showtime.value) return ''
  return `${showtime.value.hall} · ${new Date(showtime.value.start_time).toLocaleString()}`
})

async function loadSeats() {
  const { data } = await api.get(`/api/showtimes/${showtimeId}/seats`)
  seatMap.setSeats(showtimeId, data.seats)
}

async function loadShowtime() {
  const { data } = await api.get(`/api/showtimes/${showtimeId}`)
  showtime.value = data.showtime
}

async function selectSeat(seat) {
  if (seatMap.mySelectedSeatIds.includes(seat.id)) {
    await unselectSeat(seat)
    return
  }
  if (seat.status !== 'AVAILABLE') return
  try {
    const { data } = await api.post('/api/bookings/select-seat', {
      showtime_id: showtimeId,
      seat_id: seat.id,
    })
    seatMap.selectSeatLocally(seat.id, data.lock_duration_seconds)
    error.value = ''
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to select seat'
  }
}

async function unselectSeat(seat) {
  try {
    await api.post('/api/bookings/release-seat', {
      showtime_id: showtimeId,
      seat_id: seat.id,
    })
    seatMap.unselectSeatLocally(seat.id)
    error.value = ''
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to release seat'
  }
}

async function proceedToPayment() {
  try {
    const { data } = await api.post('/api/bookings', {
      showtime_id: showtimeId,
      seat_ids: seatMap.mySelectedSeatIds,
    })
    router.push(`/booking/${data.booking_id}?expiresAt=${seatMap.lockExpiresAt}`)
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to create booking'
  }
}

function seatClass(seat) {
  if (seatMap.mySelectedSeatIds.includes(seat.id)) return 'seat mine'
  return `seat ${seat.status.toLowerCase()}`
}

onMounted(async () => {
  await Promise.all([loadShowtime(), loadSeats()])

  socket = new SeatSocket({
    showtimeId,
    token: auth.token,
    onEvent: (event) => seatMap.applyRemoteEvent(event),
  })
  socket.connect()

  tickInterval = setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onBeforeUnmount(() => {
  socket?.close()
  clearInterval(tickInterval)
  seatMap.reset()
})
</script>

<template>
  <div class="seat-map-page">
    <div class="page-header">
      <button type="button" class="back-btn" @click="router.push('/')">Back</button>
      <div>
        <h1>{{ pageTitle }}</h1>
        <p v-if="showtimeDetail" class="showtime-detail">{{ showtimeDetail }}</p>
      </div>
      <span class="header-spacer"></span>
    </div>
    <p v-if="error" class="error">{{ error }}</p>

    <div class="screen-wrap" aria-label="Screen position">
      <div class="screen">SCREEN</div>
    </div>

    <div class="grid">
      <div v-for="[rowLabel, seats] in rows" :key="rowLabel" class="row">
        <span class="row-label">{{ rowLabel }}</span>
        <button
          v-for="seat in seats"
          :key="seat.id"
          :class="seatClass(seat)"
          :disabled="seat.status !== 'AVAILABLE' && !seatMap.mySelectedSeatIds.includes(seat.id)"
          @click="selectSeat(seat)"
        >
          {{ seat.number }}
        </button>
      </div>
    </div>

    <div class="legend">
      <span class="seat available"></span> Available
      <span class="seat locked"></span> Locked
      <span class="seat booked"></span> Booked
      <span class="seat mine"></span> Your selection
    </div>

    <p v-if="secondsLeft !== null" class="timer">Lock expires in {{ secondsLeft }}s</p>

    <button
      class="proceed-btn"
      :disabled="seatMap.mySelectedSeatIds.length === 0"
      @click="proceedToPayment"
    >
      Proceed to payment ({{ seatMap.mySelectedSeatIds.length }} seat(s))
    </button>
  </div>
</template>

<style scoped>
.seat-map-page {
  max-width: 760px;
  margin: 40px auto;
  padding: 0 16px;
  text-align: center;
}

.page-header {
  display: grid;
  grid-template-columns: 88px 1fr 88px;
  align-items: center;
  gap: 12px;
}

.page-header h1 {
  margin: 0;
}

.showtime-detail {
  margin: 4px 0 0;
  color: #64748b;
}

.back-btn {
  justify-self: start;
  padding: 8px 14px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background: #ffffff;
  color: #334155;
  font-weight: 700;
}

.back-btn:hover {
  background: #f8fafc;
}

.screen-wrap {
  max-width: 560px;
  margin: 28px auto 24px;
}

.screen {
  height: 28px;
  border-radius: 4px;
  background: #e5e7eb;
  border: 1px solid #cbd5e1;
  color: #4b5563;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 26px;
}

.grid {
  padding: 18px 8px;
  border-top: 1px solid #e5e7eb;
}

.row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  margin-bottom: 6px;
}
.row-label {
  width: 20px;
  color: #64748b;
  font-weight: 700;
}
.seat {
  width: 34px;
  height: 34px;
  border-radius: 8px 8px 4px 4px;
  border: 1px solid transparent;
  cursor: pointer;
  font-weight: 700;
  color: #1f2937;
  transition:
    transform 0.12s ease,
    box-shadow 0.12s ease,
    background-color 0.12s ease;
}
.seat:not(:disabled):hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.14);
}
.seat.available {
  background: #d1fae5;
  border-color: #86efac;
}
.seat.locked {
  background: #fde68a;
  border-color: #fbbf24;
  cursor: not-allowed;
}
.seat.booked {
  background: #cbd5e1;
  border-color: #94a3b8;
  color: #64748b;
  cursor: not-allowed;
}
.seat.mine {
  background: #2563eb;
  border-color: #1d4ed8;
  color: white;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.18);
}
.legend {
  margin: 16px 0;
  display: flex;
  gap: 12px;
  justify-content: center;
  align-items: center;
  font-size: 14px;
}
.legend .seat {
  width: 16px;
  height: 16px;
  cursor: default;
}
.timer {
  margin: 12px 0 0;
  color: #334155;
  font-weight: 700;
}
.proceed-btn {
  margin-top: 16px;
  padding: 10px 20px;
}
.error {
  color: #d33;
}
</style>
