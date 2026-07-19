import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export const useSeatMapStore = defineStore('seatMap', () => {
  const showtimeId = ref('')
  const seats = ref([])
  const mySelectedSeatIds = ref([])
  const lockExpiresAt = ref(null)

  const selectedSeats = computed(() =>
    seats.value.filter((seat) => mySelectedSeatIds.value.includes(seat.id)),
  )

  function setSeats(nextShowtimeId, nextSeats) {
    showtimeId.value = nextShowtimeId
    seats.value = [...nextSeats].sort((a, b) => {
      const rowCompare = a.row.localeCompare(b.row)
      return rowCompare || a.number - b.number
    })
  }

  function selectSeatLocally(seatId, lockDurationSeconds) {
    if (!mySelectedSeatIds.value.includes(seatId)) {
      mySelectedSeatIds.value.push(seatId)
    }
    lockExpiresAt.value = Date.now() + lockDurationSeconds * 1000
    updateSeatStatus(seatId, 'LOCKED')
  }

  function unselectSeatLocally(seatId) {
    mySelectedSeatIds.value = mySelectedSeatIds.value.filter((id) => id !== seatId)
    updateSeatStatus(seatId, 'AVAILABLE')
    if (mySelectedSeatIds.value.length === 0) {
      lockExpiresAt.value = null
    }
  }

  function updateSeatStatus(seatId, status) {
    const seat = seats.value.find((item) => item.id === seatId)
    if (seat) seat.status = status
  }

  function applyRemoteEvent(event) {
    const seatId = event.seat_id || event.seatId
    const status = event.status
    if (!seatId || !status) return

    updateSeatStatus(seatId, status)
    if (status === 'AVAILABLE' || status === 'BOOKED') {
      mySelectedSeatIds.value = mySelectedSeatIds.value.filter((id) => id !== seatId)
    }
  }

  function reset() {
    showtimeId.value = ''
    seats.value = []
    mySelectedSeatIds.value = []
    lockExpiresAt.value = null
  }

  return {
    showtimeId,
    seats,
    mySelectedSeatIds,
    selectedSeats,
    lockExpiresAt,
    setSeats,
    selectSeatLocally,
    unselectSeatLocally,
    applyRemoteEvent,
    reset,
  }
})
