<script setup>
import { onMounted, ref } from 'vue'
import api from '@/services/api'

const showtimes = ref([])
const loading = ref(true)
const error = ref('')

async function loadShowtimes() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/api/showtimes')
    showtimes.value = data.showtimes || []
  } catch (err) {
    error.value = err.response?.data?.error || 'Failed to load showtimes'
  } finally {
    loading.value = false
  }
}

onMounted(loadShowtimes)
</script>

<template>
  <main class="show-list-page">
    <header>
      <h1>Choose a showtime</h1>
      <button type="button" @click="loadShowtimes">Refresh</button>
    </header>

    <p v-if="loading">Loading showtimes...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="showtimes.length === 0">No showtimes available. Ask an admin to create one.</p>

    <div v-else class="showtime-list">
      <article v-for="showtime in showtimes" :key="showtime.id" class="showtime-card">
        <div>
          <h2>{{ showtime.movie_name }}</h2>
          <p>{{ showtime.hall }} · {{ new Date(showtime.start_time).toLocaleString() }}</p>
          <p>{{ showtime.rows }} rows · {{ showtime.seats_per_row }} seats per row</p>
        </div>
        <router-link :to="`/shows/${showtime.id}`">Select seats</router-link>
      </article>
    </div>
  </main>
</template>

<style scoped>
.show-list-page {
  max-width: 880px;
  margin: 40px auto;
  padding: 0 16px;
}

header,
.showtime-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

h1,
h2,
p {
  margin: 0;
}

header {
  margin-bottom: 24px;
}

.showtime-list {
  display: grid;
  gap: 12px;
}

.showtime-card {
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 16px;
}

.error {
  color: #d33;
}
</style>
