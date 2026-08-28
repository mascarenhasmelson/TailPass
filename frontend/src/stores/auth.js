import { defineStore } from 'pinia'

const API_URL = import.meta.env.VITE_API_URL

// The access token lives ONLY in this store's in-memory state - never in
// localStorage/sessionStorage - so it can't be read by an XSS payload that
// merely scans browser storage. It's lost on page reload by design; init()
// silently restores a session via the httpOnly refresh cookie instead.
export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: null,
    username: null,
    setupRequired: null, // null = unknown yet, true/false once checked
    initialized: false,
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
  },

  actions: {
    async checkStatus() {
      try {
        const res = await fetch(`${API_URL}/auth/status`)
        const data = await res.json()
        this.setupRequired = !!data.setup_required
      } catch {
        // Backend unreachable - assume setup isn't required so the user
        // lands on the login screen rather than an incorrect setup screen.
        this.setupRequired = false
      }
    },

    // Called once on app boot: checks whether an admin account exists yet,
    // then tries to silently restore a session from the refresh cookie.
    async init() {
      if (this.initialized) return
      await this.checkStatus()
      await this.refresh()
      this.initialized = true
    },

    async register(username, password) {
      const res = await fetch(`${API_URL}/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username, password }),
      })
      if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw new Error(text || 'Registration failed')
      }
      const data = await res.json()
      this._applySession(data)
      this.setupRequired = false
    },

    async login(username, password) {
      const res = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ username, password }),
      })
      if (!res.ok) {
        const text = await res.text().catch(() => '')
        throw new Error(text || 'Invalid username or password')
      }
      const data = await res.json()
      this._applySession(data)
    },

    // Exchanges the httpOnly refresh cookie for a fresh access token.
    // Returns true on success, false if there's no valid session.
    async refresh() {
      try {
        const res = await fetch(`${API_URL}/auth/refresh`, {
          method: 'POST',
          credentials: 'include',
        })
        if (!res.ok) {
          this.accessToken = null
          this.username = null
          return false
        }
        const data = await res.json()
        this._applySession(data)
        return true
      } catch {
        this.accessToken = null
        this.username = null
        return false
      }
    },

    async logout() {
      try {
        await fetch(`${API_URL}/auth/logout`, {
          method: 'POST',
          credentials: 'include',
        })
      } catch {
        // Ignore network errors - clear local state regardless below.
      }
      this.accessToken = null
      this.username = null
    },

    _applySession(data) {
      this.accessToken = data.access_token
      this.username = data.username
    },

    // Authenticated fetch helper: attaches the bearer token, and if the
    // server reports the token expired (401), silently refreshes once and
    // retries before giving up.
    async authFetch(path, options = {}) {
      const doFetch = () =>
        fetch(`${API_URL}${path}`, {
          ...options,
          headers: {
            ...(options.headers || {}),
            ...(this.accessToken ? { Authorization: `Bearer ${this.accessToken}` } : {}),
          },
          credentials: 'include',
        })

      let res = await doFetch()
      if (res.status === 401 && this.accessToken) {
        const refreshed = await this.refresh()
        if (refreshed) {
          res = await doFetch()
        }
      }
      return res
    },
  },
})
