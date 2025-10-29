<template>
  <div class="app-wrapper">
    <div class="app-container">
      <div class="header">
        <h2>Service Dashboard</h2>
        <div>
          <button class="small-btn" @click="toggleDark">
            {{ isDark ? 'Light' : 'Dark' }} Mode
          </button>
          <button @click="showForm = true">+ Add Service</button>
        </div>
      </div>

      <!-- Service list -->
      <div class="service-list">
        <div v-for="service in services" :key="service.id" class="service-card">
          <div>
            <div><strong>{{ service.service_name }}</strong></div>
            <div style="color:var(--muted); font-size:0.9rem;">
              Local: {{ service.local_ip }}:{{ service.local_port }}
              • Remote: {{ service.remote_ip }}:{{ service.remote_port }}
            </div>
          </div>
          <div>
            <button class="small-btn" @click="deleteService(service.id)">Delete</button>
          </div>
        </div>
      </div>

      <!-- Modal -->
      <div v-if="showForm" class="modal" @click.self="showForm = false">
        <div class="modal-content">
          <h3>Add Service</h3>
          <form @submit.prevent="saveService">
            <input v-model="form.service_name" placeholder="Service Name" required />
            <input v-model="form.local_ip" placeholder="Local IP" required />
            <input v-model="form.local_port" placeholder="Local Port" type="number" required />
            <input v-model="form.remote_ip" placeholder="Remote IP" required />
            <input v-model="form.remote_port" placeholder="Remote Port" type="number" required />
            <div style="display:flex; gap:8px; justify-content:flex-end; margin-top:8px;">
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
import { ref, onMounted } from 'vue';

const showForm = ref(false);
const isDark = ref(false);

const services = ref([]); // example; replace with fetched data
const form = ref({
  service_name: '',
  local_ip: '',
  local_port: '',
  remote_ip: '',
  remote_port: ''
});

// load preference on mount
onMounted(() => {
  const stored = localStorage.getItem('dark-mode');
  if (stored !== null) {
    isDark.value = stored === 'true';
  } else {
    // optional: respect system preference
    isDark.value = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
  }
  applyBodyClass();
});

// toggle function
function toggleDark() {
  isDark.value = !isDark.value;
  localStorage.setItem('dark-mode', isDark.value);
  applyBodyClass();
}

function applyBodyClass() {
  if (isDark.value) document.body.classList.add('dark');
  else document.body.classList.remove('dark');
}

// dummy save (replace with API call)
function saveService() {
  if (!form.value.service_name.trim()) {
    alert('Service name required');
    return;
  }
  // push a copy
  services.value.push({ ...form.value, id: Date.now() });
  // reset
  form.value = { service_name: '', local_ip: '', local_port: '', remote_ip: '', remote_port: '' };
  showForm.value = false;
}

// dummy delete (replace with API delete)
function deleteService(id) {
  const idx = services.value.findIndex(s => s.id === id);
  if (idx !== -1) services.value.splice(idx, 1);
}
</script>

<style>
:root {
  /* Light theme */
  --bg: #f5f7fb;
  --card-bg: #ffffff;
  --text: #111827;
  --muted: #6b7280;
  --accent: #2563eb;
  --modal-overlay: rgba(0,0,0,0.45);
}

body.dark {
  /* Dark theme */
  --bg: #0b1220;
  --card-bg: #0f1724;
  --text: #e6eef8;
  --muted: #9aa6b2;
  --accent: #4f8ef7;
  --modal-overlay: rgba(0,0,0,0.7);
}

html, body {
  height: 100%;
  margin: 0;
}

.app-wrapper {
  min-height: 100%;
  display: flex;
  justify-content: flex-end; /* if you want container at right */
  align-items: flex-start;
  background: var(--bg);
  color: var(--text);
  font-family: Arial, sans-serif;
}

/* container */
.app-container {
  width: 420px;
  padding: 20px;
  box-sizing: border-box;
}

/* header area with toggle */
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

/* cards */
.service-list { margin-top: 12px; }
.service-card {
  background: var(--card-bg);
  color: var(--text);
  border-radius: 8px;
  padding: 12px;
  box-shadow: 0 4px 10px rgba(2,6,23,0.08);
  margin-bottom: 10px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* buttons */
button {
  padding: 6px 10px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  background: var(--accent);
  color: white;
  font-weight: 600;
}
.small-btn {
  padding: 4px 8px;
  font-size: 0.9rem;
}

/* modal */
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
}
.modal-content {
  width: 320px;
  background: var(--card-bg);
  color: var(--text);
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 10px 30px rgba(2,6,23,0.25);
}
input {
  width: 100%;
  margin-bottom: 8px;
  padding: 8px;
  border-radius: 6px;
  border: 1px solid #cbd5e1;
  box-sizing: border-box;
  background: transparent;
  color: var(--text);
}
body.dark input {
  border-color: #233043;
}

</style>
