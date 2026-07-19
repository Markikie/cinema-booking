
export class SeatSocket {
    constructor({ showtimeId, token, onEvent }) {
      this.showtimeId = showtimeId
      this.token = token
      this.onEvent = onEvent
      this.ws = null
      this.closedByClient = false
      this.retryDelay = 1000
    }
  
    connect() {
      const apiBase = import.meta.env.VITE_API_BASE_URL || window.location.origin
      const defaultBase = apiBase.replace(/^http/, 'ws')
      const base = import.meta.env.VITE_WS_BASE_URL || defaultBase
      const url = `${base}/ws?showtime_id=${encodeURIComponent(this.showtimeId)}&token=${encodeURIComponent(this.token)}`
  
      this.ws = new WebSocket(url)
  
      this.ws.onmessage = (msg) => {
        try {
          const event = JSON.parse(msg.data)
          this.onEvent(event)
        } catch (err) {
          console.error('failed to parse seat event', err)
        }
      }
  
      this.ws.onclose = () => {
        if (this.closedByClient) return
        setTimeout(() => this.connect(), this.retryDelay)
        this.retryDelay = Math.min(this.retryDelay * 2, 15000)
      }
  
      this.ws.onopen = () => {
        this.retryDelay = 1000
      }
    }
  
    close() {
      this.closedByClient = true
      this.ws?.close()
    }
  }
  
