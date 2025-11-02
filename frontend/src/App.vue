<template>
  <div class="app-wrapper">
    <div class="app-container">
      <div class="header">
      <div class="h1">
        <h2>TailPass</h2>
      </div>

        <div class="button-group">
          <button @click="toggleDark">{{ isDark ? 'Light Mode' : 'Dark Mode' }}</button>
          <button @click="showForm = true"> Add Service</button>
        </div>
      </div>

      <div class="service-list">
        <div v-for="service in services" :key="service.id" class="service-card">
          <div class="service-info">
            <div class="service-name"><strong>{{ service.service_name }}</strong></div>
            <div class="service-details">
               <!-- <div>
                <span class="label">id:</span>
                <span class="highlight-ip">{{ service.id }}</span>
              </div> -->
              <div>
                <span class="label">Local IP:</span>
                <span class="highlight-ip">{{ service.local_ip }}</span>
              </div>
              <div>
                <span class="label">Local Port:</span>
                <span class="highlight-port">{{ service.local_port }}</span>
              </div>
              <div>
                <span class="label">Remote IP:</span>
                <span class="highlight-ip">{{ service.remote_ip }}</span>
              </div>
              <div>
                <span class="label">Remote Port:</span>
                <span class="highlight-port">{{ service.remote_port }}</span>
              </div>
            </div>
          </div>
          <div>
            <button class="delete" @click="deleteService(service.id)">Delete</button>
          </div>
        </div>
      </div>

      <div v-if="showForm" class="modal" @click.self="showForm = false">
        <div class="modal-content">
          <h3>Add Service</h3>
          <form @submit.prevent="saveService">
            <input v-model="form.service_name" placeholder="Service Name" required />
            <input v-model="form.local_ip" placeholder="Local IP" required />
            <input v-model="form.local_port" placeholder="Local Port" type="number" required />
            <input v-model="form.remote_ip" placeholder="Remote IP" required />
            <input v-model="form.remote_port" placeholder="Remote Port" type="number" required />
            <div class="modal-actions">
              <button type="button" @click="showForm = false">Cancel</button>
              <button type="submit">Save</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted ,onUnmounted} from 'vue';
const API_URL = import.meta.env.VITE_API_URL;
const showForm = ref(false);
const isDark = ref(false);
const loading = ref(false);
const error = ref(null);

const services = ref([]);
const form = ref({
  service_name: '',
  local_ip: '',
  local_port: '',
  remote_ip: '',
  remote_port: ''
});

async function fetchServices() {
  loading.value = true;
  error.value = null;

  try {
  const response = await fetch(`${API_URL}/services`);

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    services.value = data;
  } catch (err) {
    error.value = `Failed to fetch services: ${err.message}`;
    console.error('Error fetching services:', err);
  } finally {
    loading.value = false;
  }
}

async function saveService() {
  if (!form.value.service_name.trim()) {
    alert('Service name is required');
    return;
  }

  loading.value = true;
  error.value = null;

  try {
    const response = await fetch(`${API_URL}/services`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    body: JSON.stringify(form.value)
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    form.value = {
      service_name: '',
      local_ip: '',
      local_port: '',
      remote_ip: '',
      remote_port: ''
    };
    showForm.value = false;

    await fetchServices();
  } catch (err) {
    error.value = `Failed to save service: ${err.message}`;
    console.error('Error saving service:', err);
    alert(`Error: ${err.message}`);
  } finally {
    loading.value = false;
  }
}
async function deleteService(id) {
  if (!confirm('delete?')) {
    return;
  }

  loading.value = true;
  error.value = null;

  try {
    const response = await fetch(`${API_URL}/services/${id}`, {
      method: 'DELETE'
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    await fetchServices();
  } catch (err) {
    error.value = `Failed to delete service: ${err.message}`;
    console.error('Error deleting service:', err);
    alert(`Error: ${err.message}`);
  } finally {
    loading.value = false;
  }
}

let intervalId = null;

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId);
});

onMounted(() => {
  const stored = localStorage.getItem('dark-mode');
  if (stored !== null) {
    isDark.value = stored === 'true';
  } else {
    isDark.value = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
  }
  applyBodyClass();

  fetchServices();
   intervalId = setInterval(fetchServices, 5000);
});

function toggleDark() {
  isDark.value = !isDark.value;
  localStorage.setItem('dark-mode', isDark.value);
  applyBodyClass();
}

function applyBodyClass() {
  if (isDark.value) document.body.classList.add('dark');
  else document.body.classList.remove('dark');
}
</script>

<style>
html, body {
  min-height: 100vh;
  width: 100vw;
  margin: 0;
  padding: 0;
  background: var(--bg);
  color: var(--text);
  transition: background 0.3s, color 0.3s;
}

:root {
  --bg: #f5f7fb;
  --card-bg: #ffffff;
  --text: #111827;
  --muted: #6b7280;
  --btn-bg: #2563eb;
  --btn-text: #ffffff;
  --modal-overlay: rgba(0, 0, 0, 0.5);
}

body.dark {
  --bg: #0b1220;
  --card-bg: #0f1724;
  --text: #e6eef8;
  --muted: #9ca3af;
  --btn-bg: #3b82f6;
  --btn-text: #ffffff;
  --modal-overlay: rgba(0, 0, 0, 0.7);
}
.app-wrapper, .app-container {
  width: 100%;
  min-height: 100vh;
  background: var(--bg);
  box-sizing: border-box;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.button-group {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: flex-end;
}
.service-list {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  width: 100%;
  margin-top: 20px;
}

.service-card {
  width: 100%;
  max-width: 900px;
  background: var(--card-bg);
  color: var(--text);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.2s;
}

.service-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.service-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.service-name {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--text);
}

.service-details {
  color: var(--muted);
  font-size: 0.95rem;
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 20px;
  line-height: 1.6;
  white-space: nowrap;
}

.service-details > div {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.service-details .label {
  font-weight: 500;
  color: var(--text);
  opacity: 0.7;
  margin-right: 4px;
}

.highlight-ip {
  font-weight: 600;
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
}

.highlight-port {
  font-weight: 600;
  color: #10b981;
  background: rgba(16, 185, 129, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
}

body.dark .highlight-ip {
  color: #60a5fa;
  background: rgba(96, 165, 250, 0.15);
}

body.dark .highlight-port {
  color: #34d399;
  background: rgba(52, 211, 153, 0.15);
}

button {
  padding: 8px 14px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.3s;
}

.button-group button:first-child {
  background: transparent;
  border: 1px solid var(--text);
  color: var(--text);
}

.button-group button:last-child {
  background: #3b82f6;
  color: white;
}

button:hover {
  opacity: 0.9;
}

button.delete {
  background-color: #dc2626;
  color: white;
  flex-shrink: 0;
  margin-left: 24px;
  padding: 10px 16px;
  white-space: nowrap;
}

button.delete:hover {
  background-color: #b91c1c;
}
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: var(--modal-overlay);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  width: 400px;
  background: var(--card-bg);
  color: var(--text);
  border-radius: 12px;
  padding: 24px;
  box-shadow:
    0 10px 30px rgba(2, 6, 23, 0.25),
    0 0 0 1px rgba(59, 130, 246, 0.3),
    0 0 40px rgba(59, 130, 246, 0.4),
    0 0 80px rgba(59, 130, 246, 0.2);
  animation: modalGlow 0.3s ease-out;
}

.modal-content h3 {
  margin-top: 0;
  margin-bottom: 16px;
}

@keyframes modalGlow {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 16px;
}
input {
  width: 100%;
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 6px;
  border: 1px solid #cbd5e1;
  box-sizing: border-box;
  background: transparent;
  color: var(--text);
  font-size: 0.95rem;
}

input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

body.dark input {
  border-color: #233043;
}

body.dark input:focus {
  border-color: #3b82f6;
}
</style>
