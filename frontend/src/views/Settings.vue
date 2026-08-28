<template>
  <main class="settings-page">
    <header class="settings-header">
      <h1>Settings</h1>
      <p class="subtitle">Connection details and network status for this TailPass node</p>
    </header>

    <section class="settings-grid">
      <div class="settings-card">
        <div class="card-title">
          <span class="icon">🛰️</span>
          <span>Tailscale IP</span>
        </div>
        <div class="card-body">
          <span v-if="tsLoading" class="value muted">Detecting…</span>
          <span v-else-if="tsError" class="value error">Not detected</span>
          <span v-else class="value mono">{{ tailscaleIP }}</span>
          <p class="hint" v-if="tsError">
            {{ tsError }} — make sure Tailscale is installed and connected on this host.
          </p>
          <p class="hint" v-else>
            New services always bind to this address by default, so tunnels stay
            reachable over your tailnet without typing an IP manually.
          </p>
        </div>
        <button class="refresh-btn" @click="fetchTailscaleIP" :disabled="tsLoading">
          Refresh
        </button>
      </div>

      <div class="settings-card">
        <div class="card-title">
          <span class="icon">🔌</span>
          <span>Backend API</span>
        </div>
        <div class="card-body">
          <span class="value mono">{{ apiUrl }}</span>
          <p class="hint">Configured via the <code>VITE_API_URL</code> environment variable.</p>
        </div>
      </div>

      <div class="settings-card">
        <div class="card-title">
          <span class="icon">🎲</span>
          <span>Local Port Allocation</span>
        </div>
        <div class="card-body">
          <span class="value">Automatic</span>
          <p class="hint">
            When you add a service without picking a local port, TailPass asks the OS
            for a free, unused port on the Tailscale interface and uses that.
          </p>
        </div>
      </div>

      <div class="settings-card">
        <div class="card-title">
          <span class="icon">🧵</span>
          <span>Tunnel Engine</span>
        </div>
        <div class="card-body">
          <span class="value">Built-in (in-process)</span>
          <p class="hint">
            TCP forwarding runs natively inside the TailPass backend — no external
            tunnel binary is spawned or required.
          </p>
        </div>
      </div>
    </section>
  </main>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const API_URL = import.meta.env.VITE_API_URL
const apiUrl = ref(API_URL)
const tailscaleIP = ref('')
const tsLoading = ref(true)
const tsError = ref('')

const auth = useAuthStore()

async function fetchTailscaleIP() {
  tsLoading.value = true
  tsError.value = ''
  try {
    const response = await auth.authFetch('/services/tailscale-ip')
    const data = await response.json()
    if (!response.ok) {
      throw new Error(data.Message || 'Failed to detect Tailscale IP')
    }
    tailscaleIP.value = data.ip
  } catch (err) {
    tsError.value = err.message
  } finally {
    tsLoading.value = false
  }
}

onMounted(fetchTailscaleIP)
</script>

<style scoped>
.settings-page {
  background: #0f172a;
  min-height: 100vh;
  padding: 40px;
  color: #e2e8f0;
}

.settings-header h1 {
  margin: 0;
  font-size: 2.2rem;
  font-weight: 700;
}

.subtitle {
  margin: 8px 0 0;
  color: #94a3b8;
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
  margin-top: 32px;
}

.settings-card {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 14px;
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 1.05rem;
}

.card-title .icon {
  font-size: 1.4rem;
}

.card-body .value {
  display: block;
  font-size: 1.3rem;
  font-weight: 700;
  color: #60a5fa;
}

.card-body .value.mono {
  font-family: 'Courier New', monospace;
}

.card-body .value.muted {
  color: #94a3b8;
  font-weight: 500;
  font-size: 1rem;
}

.card-body .value.error {
  color: #f87171;
}

.hint {
  margin: 8px 0 0;
  font-size: 0.85rem;
  color: #94a3b8;
  line-height: 1.5;
}

.refresh-btn {
  align-self: flex-start;
  background: rgba(96, 165, 250, 0.15);
  color: #93c5fd;
  border: 1px solid rgba(96, 165, 250, 0.3);
  padding: 8px 16px;
  border-radius: 8px;
  font-weight: 600;
  transition: all 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: rgba(96, 165, 250, 0.25);
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
