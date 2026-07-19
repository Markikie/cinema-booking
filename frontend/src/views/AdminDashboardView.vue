<script setup>
import { onMounted, reactive, ref } from 'vue'
import api from '@/services/api'

const bookings = ref([])
const auditLogs = ref([])
const error = ref('')


const filters = reactive({ showtime_id: '', status: '' })

const newShowtime = reactive({
  movie_name: '',
  hall: '',
  start_time: '',
  rows: 5,
  seats_per_row: 8,
})
const showtimeCreateMessage = ref('')

async function loadBookings() {
  const params = {}
  if (filters.showtime_id) params.showtime_id = filters.showtime_id
  if (filters.status) params.status = filters.status
  try {
    const { data } = await api.get('/api/admin/bookings', { params })
    bookings.value = data.bookings
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load bookings'
  }
}

async function loadAuditLogs() {
  try {
    const { data } = await api.get('/api/admin/audit-logs')
    auditLogs.value = data.logs
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load audit logs'
  }
}

async function createShowtime() {
  showtimeCreateMessage.value = ''
  try {
    const { data } = await api.post('/api/admin/showtimes', {
      ...newShowtime,
      rows: Number(newShowtime.rows),
      seats_per_row: Number(newShowtime.seats_per_row),
      start_time: new Date(newShowtime.start_time).toISOString(),
    })
    showtimeCreateMessage.value = `Created ${data.showtime?.movie_name || newShowtime.movie_name} with ${data.seats_created} seats.`
  } catch (err) {
    showtimeCreateMessage.value = err.response?.data?.error || 'Failed to create showtime'
  }
}

onMounted(() => {
  loadBookings()
  loadAuditLogs()
})
</script>

<template>
  <div class="admin-page">
    <h1>Admin Dashboard</h1>
    <p v-if="error" class="error">{{ error }}</p>

    <section>
      <h2>Create Showtime (seeds seat map)</h2>
      <form class="showtime-form" @submit.prevent="createShowtime">
        <input v-model="newShowtime.movie_name" placeholder="Movie name" required />
        <input v-model="newShowtime.hall" placeholder="Hall" required />
        <input v-model="newShowtime.start_time" type="datetime-local" required />
        <input v-model="newShowtime.rows" type="number" min="1" placeholder="Rows" required />
        <input v-model="newShowtime.seats_per_row" type="number" min="1" placeholder="Seats per row" required />
        <button type="submit">Create</button>
      </form>
      <p v-if="showtimeCreateMessage">{{ showtimeCreateMessage }}</p>
    </section>

    <section>
      <h2>Bookings</h2>
      <div class="filters">
        <input v-model="filters.showtime_id" placeholder="Filter by showtime_id" />
        <select v-model="filters.status">
          <option value="">All statuses</option>
          <option value="PENDING">PENDING</option>
          <option value="SUCCESS">SUCCESS</option>
          <option value="TIMEOUT">TIMEOUT</option>
          <option value="FAILED">FAILED</option>
        </select>
        <button @click="loadBookings">Apply</button>
      </div>
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>User</th>
            <th>Showtime</th>
            <th>Seats</th>
            <th>Status</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in bookings" :key="b.id">
            <td>{{ b.id }}</td>
            <td>{{ b.user_id }}</td>
            <td>{{ b.showtime_id }}</td>
            <td>{{ b.seat_ids.join(', ') }}</td>
            <td>{{ b.status }}</td>
            <td>{{ new Date(b.created_at).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </section>

    <section>
      <h2>Audit Logs</h2>
      <table>
        <thead>
          <tr>
            <th>Event</th>
            <th>User</th>
            <th>Booking</th>
            <th>Detail</th>
            <th>Time</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in auditLogs" :key="log.id">
            <td>{{ log.event_type }}</td>
            <td>{{ log.user_id }}</td>
            <td>{{ log.booking_id }}</td>
            <td>{{ log.detail }}</td>
            <td>{{ new Date(log.created_at).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>

<style scoped>
.admin-page {
  max-width: 960px;
  margin: 40px auto;
  padding: 0 16px;
}
section {
  margin-bottom: 40px;
}
.showtime-form,
.filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}
table {
  width: 100%;
  border-collapse: collapse;
}
th,
td {
  border: 1px solid #ddd;
  padding: 6px 8px;
  font-size: 14px;
  text-align: left;
}
.error {
  color: #d33;
}
</style>
