<template>
  <div class="app-wrapper">
    <div class="app-container">
      <!-- Header with stats -->
      <header class="dashboard-header">
        <div class="header-content">
          <h1>Service Monitoring</h1>
          <p class="subtitle">Monitor your port forwarding services in real-time</p>
        </div>
        <div class="header-stats">
          <div class="stat-card">
            <div class="stat-icon">📊</div>
            <div class="stat-info">
              <span class="stat-value">{{ services.length }}</span>
              <span class="stat-label">Total Services</span>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">✅</div>
            <div class="stat-info">
              <span class="stat-value">{{ onlineCount }}</span>
              <span class="stat-label">Online</span>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">❌</div>
            <div class="stat-info">
              <span class="stat-value">{{ offlineCount }}</span>
              <span class="stat-label">Offline</span>
            </div>
          </div>
        </div>
      </header>

      <!-- Action Bar -->
      <div class="action-bar">
        <div class="search-filter">
          <input 
            v-model="searchQuery" 
            placeholder="Search services..." 
            class="search-input"
          >
          <div class="filter-group">
            <button 
              :class="['filter-btn', activeFilter === 'all' ? 'active' : '']"
              @click="activeFilter = 'all'"
            >
              All ({{ services.length }})
            </button>
            <button 
              :class="['filter-btn', activeFilter === 'online' ? 'active' : '']"
              @click="activeFilter = 'online'"
            >
              Online ({{ onlineCount }})
            </button>
            <button 
              :class="['filter-btn', activeFilter === 'offline' ? 'active' : '']"
              @click="activeFilter = 'offline'"
            >
              Offline ({{ offlineCount }})
            </button>
          </div>
        </div>
        <button class="add-btn" @click="showForm = true">
          <span class="btn-icon">+</span>
          Add New Service
        </button>
      </div>

      <!-- Loading State - Show only when initially loading -->
      <div v-if="loading && services.length === 0" class="loading-overlay">
        <div class="loading-spinner"></div>
        <p>Loading services...</p>
      </div>

      <!-- Services Grid - Show when not loading and there are services -->
      <div v-else-if="filteredServices.length > 0" class="services-grid">
        <div 
          v-for="service in filteredServices" 
          :key="service.id" 
          :class="['service-card', service.online ? 'online' : 'offline']"
        >
          <div class="service-header">
            <div class="service-title">
              <div class="service-icon">
                <span v-if="service.online">🌐</span>
                <span v-else>🔴</span>
              </div>
              <div>
                <h3>{{ service.service_name }}</h3>
                <div class="service-status">
                  <span :class="['status-badge', service.online ? 'online' : 'offline']">
                    {{ service.online ? 'Online' : 'Offline' }}
                  </span>
                </div>
              </div>
            </div>
            <div class="service-actions">
              <!-- <button 
                class="action-btn refresh" 
                @click="refreshService(service.id)"
                title="Refresh status"
              >
                ↻
              </button> -->
              <button 
                class="action-btn delete" 
                @click="confirmDelete(service.id)"
                title="Delete service"
              >
                ×
              </button>
            </div>
          </div>

          <div class="service-details">
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">Local Endpoint</span>
                <div class="detail-value endpoint">
                  <span class="ip-address">{{ service.local_ip }}</span>
                  <span class="port-badge">:{{ service.local_port }}</span>
                </div>
              </div>
              <div class="detail-item">
                <span class="detail-label">Remote Endpoint</span>
                <div class="detail-value endpoint">
                  <span class="ip-address">{{ service.remote_ip }}</span>
                  <span class="port-badge">:{{ service.remote_port }}</span>
                </div>
              </div>
            </div>

            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">Last Seen</span>
                <div class="detail-value">
                  <span :class="['last-seen', service.online ? 'online' : 'offline']">
                    {{ formatLastSeen(service.last_seen) }}
                  </span>
                </div>
              </div>
              <div class="detail-item">
                <span class="detail-label">Protocol</span>
                <div class="detail-value">
                  <span class="protocol-badge">TCP</span>
                </div>
              </div>
            </div>

            <div class="service-footer">
              <div class="uptime-indicator">
                <div class="uptime-bar">
                  <div 
                    :class="['uptime-fill', service.online ? 'online' : 'offline']"
                    :style="{ width: service.online ? '100%' : '0%' }"
                  ></div>
                </div>
                <span class="uptime-text">
                  {{ service.online ? '100% Uptime' : 'Service Down' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State - Show when not loading and there are no services -->
      <div v-else class="empty-state">
        <div class="empty-icon">📡</div>
        <h3>No services found</h3>
        <p v-if="searchQuery">No services match your search criteria</p>
        <p v-else>Add your first service to start monitoring</p>
      </div>

      <!-- Error State -->
      <div v-if="error" class="error-banner">
        <div class="error-content">
          <span class="error-icon">⚠️</span>
          <span>{{ error }}</span>
        </div>
        <button class="retry-btn" @click="fetchServices">Retry</button>
      </div>

      <!-- Add Service Modal -->
      <div v-if="showForm" class="modal-overlay" @click.self="closeModal">
        <div class="modal">
          <div class="modal-header">
            <h2>Add New Service</h2>
            <button class="close-btn" @click="closeModal">×</button>
          </div>
          <div class="modal-content">
            <form @submit.prevent="saveService">
              <div class="form-group">
                <label for="service_name">Service Name *</label>
                <input 
                  id="service_name"
                  v-model="form.service_name" 
                  placeholder="e.g., Web Server, SSH Tunnel" 
                  required
                />
              </div>
              
              <div class="form-row">
                <div class="form-group">
                  <label for="local_ip">Local IP *</label>
                  <input 
                    id="local_ip"
                    v-model="form.local_ip" 
                    placeholder="192.168.1.100" 
                    required
                  />
                </div>
                <div class="form-group">
                  <label for="local_port">Local Port *</label>
                  <input 
                    id="local_port"
                    v-model="form.local_port" 
                    placeholder="80" 
                    type="number"
                    min="1"
                    max="65535"
                    required
                  />
                </div>
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label for="remote_ip">Remote IP *</label>
                  <input 
                    id="remote_ip"
                    v-model="form.remote_ip" 
                    placeholder="10.0.0.1" 
                    required
                  />
                </div>
                <div class="form-group">
                  <label for="remote_port">Remote Port *</label>
                  <input 
                    id="remote_port"
                    v-model="form.remote_port" 
                    placeholder="8080" 
                    type="number"
                    min="1"
                    max="65535"
                    required
                  />
                </div>
              </div>

              <div class="form-note">
                <span class="note-icon">💡</span>
                <span>All fields marked with * are required. Service status will be checked automatically.</span>
              </div>

              <div class="form-actions">
                <button type="button" class="cancel-btn" @click="closeModal">Cancel</button>
                <button type="submit" class="submit-btn" :disabled="saving">
                  <span v-if="saving" class="spinner"></span>
                  {{ saving ? 'Saving...' : 'Save Service' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- Delete Confirmation Modal -->
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = null">
        <div class="confirm-modal">
          <div class="confirm-icon">🗑️</div>
          <h3>Delete Service</h3>
          <p>Are you sure you want to delete this service? This action cannot be undone.</p>
          <div class="confirm-actions">
            <button class="cancel-btn" @click="showDeleteConfirm = null">Cancel</button>
            <button class="delete-confirm-btn" @click="deleteService(showDeleteConfirm)">
              Delete Service
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';

//const API_URL = "http://192.168.20.17:8082";

const API_URL = import.meta.env.VITE_API_URL;
const showForm = ref(false);
const showDeleteConfirm = ref(null);
const loading = ref(true); // Start with loading true
const saving = ref(false);
const error = ref(null);
const searchQuery = ref('');
const activeFilter = ref('all');
const services = ref([]);
const initialLoad = ref(true); // Track if this is the first load

const form = ref({
  service_name: '',
  local_ip: '',
  local_port: '',
  remote_ip: '',
  remote_port: ''
});

const onlineCount = computed(() => 
  services.value.filter(s => s.online).length
);

const offlineCount = computed(() => 
  services.value.filter(s => !s.online).length
);

const filteredServices = computed(() => {
  let filtered = services.value;
  
  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(service => 
      service.service_name.toLowerCase().includes(query) ||
      service.local_ip.toLowerCase().includes(query) ||
      service.remote_ip.toLowerCase().includes(query)
    );
  }
  
  // Apply status filter
  if (activeFilter.value === 'online') {
    filtered = filtered.filter(service => service.online);
  } else if (activeFilter.value === 'offline') {
    filtered = filtered.filter(service => !service.online);
  }
  
  return filtered;
});

async function fetchServices() {
  // Don't show loading overlay for periodic refreshes
  if (initialLoad.value) {
    loading.value = true;
  }
  
  error.value = null;

  try {
    const response = await fetch(`${API_URL}/services`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    services.value = data || []; // Ensure it's always an array
  } catch (err) {
    error.value = `Failed to fetch services: ${err.message}`;
    console.error('Error fetching services:', err);
    services.value = []; // Set to empty array on error
  } finally {
    loading.value = false;
    initialLoad.value = false; // Mark initial load as complete
  }
}

async function refreshService(id) {
  const service = services.value.find(s => s.id === id);
  if (!service) return;
  
  // Optimistically update UI
  service.online = false;
  
  try {
    const response = await fetch(`${API_URL}/services/${id}/check`);
    if (response.ok) {
      const data = await response.json();
      service.online = data.online;
      service.last_seen = data.last_seen;
    }
  } catch (err) {
    console.error('Error refreshing service:', err);
  }
}

function confirmDelete(id) {
  showDeleteConfirm.value = id;
}

async function deleteService(id) {
  if (!showDeleteConfirm.value) return;

  try {
    const response = await fetch(`${API_URL}/services/${id}`, {
      method: 'DELETE'
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    services.value = services.value.filter(service => service.id !== id);
    showDeleteConfirm.value = null;
  } catch (err) {
    error.value = `Failed to delete service: ${err.message}`;
    console.error('Error deleting service:', err);
  }
}

function formatLastSeen(lastSeen) {
  if (!lastSeen) return "Never";
  const date = new Date(lastSeen);
  if (isNaN(date.getTime())) return "Invalid date";
  
  const now = new Date();
  const diffMs = now - date;
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);
  
  if (diffSec < 60) return "Just now";
  if (diffMin < 60) return `${diffMin} minute${diffMin > 1 ? 's' : ''} ago`;
  if (diffHour < 24) return `${diffHour} hour${diffHour > 1 ? 's' : ''} ago`;
  if (diffDay < 7) return `${diffDay} day${diffDay > 1 ? 's' : ''} ago`;
  
  return date.toLocaleDateString();
}

async function saveService() {
  if (!form.value.service_name.trim()) {
    alert('Service name is required');
    return;
  }

  saving.value = true;

  try {
    const response = await fetch(`${API_URL}/services`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        service_name: form.value.service_name,
        local_ip: form.value.local_ip,
        local_port: form.value.local_port,
        remote_ip: form.value.remote_ip,
        remote_port: form.value.remote_port
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    // Reset form
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
    saving.value = false;
  }
}

function closeModal() {
  showForm.value = false;
  form.value = {
    service_name: '',
    local_ip: '',
    local_port: '',
    remote_ip: '',
    remote_port: ''
  };
}

let intervalId = null;

onMounted(() => {
  fetchServices();
  intervalId = setInterval(fetchServices, 10000); // Refresh every 10 seconds
});

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId);
});
</script>

<style scoped>
/* FIXED: Full width layout */
.app-wrapper {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  padding: 0; /* Removed padding to allow full width */
  width: 100vw; /* Ensure full viewport width */
  margin: 0; /* Remove any margin */
  overflow-x: hidden; /* Prevent horizontal scroll */
}

/* FIXED: Container takes full width */
.app-container {
  width: 100%; /* Full width instead of max-width */
  min-height: 100vh; /* Full viewport height */
  background: rgba(15, 23, 42, 0.8);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(148, 163, 184, 0.1);
  margin: 0; /* Remove auto margin */
  border-radius: 0; /* Remove border radius for full width */
}

/* Header */
.dashboard-header {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(15, 23, 42, 0.9) 100%);
  color: #e2e8f0;
  padding: 40px 40px 30px;
  position: relative;
  overflow: hidden;
  width: 100%;
  border-bottom: 3px solid #3b82f6;
}

.dashboard-header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.1), transparent);
  animation: shimmer 3s infinite;
}

@keyframes shimmer {
 0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.header-content h1 {
  margin: 0;
  font-size: 2.5rem;
  font-weight: 700;
  position: relative;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}


.subtitle {
  margin: 8px 0 0;
  opacity: 0.9;
  font-size: 1.1rem;
  position: relative;
  color: rgba(255, 255, 255, 0.9);
}

.header-stats {
  display: flex;
  gap: 20px;
  margin-top: 30px;
  position: relative;
  max-width: 1400px;
  margin: 30px auto 0;
  width: 100%;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 15px;
  background: rgba(255, 255, 255, 0.1);
  padding: 20px;
  border-radius: 12px;
  backdrop-filter: blur(10px);
  flex: 1;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s;
}


.stat-card:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

.stat-icon {
  font-size: 2rem;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.stat-value {
  display: block;
  font-size: 2rem;
  font-weight: 700;
  color: white;
}

.stat-label {
  display: block;
  opacity: 0.9;
  font-size: 0.9rem;
  color: rgba(255, 255, 255, 0.8);
}

/* Action Bar */
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 25px 40px;
  background: rgba(30, 41, 59, 0.8);
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

.search-filter {
  display: flex;
  gap: 20px;
  align-items: center;
  flex: 1;
}

.search-input {
  padding: 12px 20px;
  border: 2px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  font-size: 1rem;
  width: 300px;
  transition: all 0.3s;
  background: rgba(15, 23, 42, 0.7);
  color: #e2e8f0;
  backdrop-filter: blur(10px);
}

.search-input::placeholder {
  color: #94a3b8;
}

.search-input:focus {
  outline: none;
  border-color: #84080c;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  background: rgba(15, 23, 42, 0.9);
}

.filter-group {
  display: flex;
  gap: 10px;
}

.filter-btn {
  padding: 8px 16px;
  border: 2px solid rgba(148, 163, 184, 0.2);
  background: rgba(30, 41, 59, 0.8);
  color: #cbd5e1;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.filter-btn:hover {
  border-color: #60a5fa;
  background: rgba(30, 41, 59, 0.9);
}

.filter-btn.active {
  background: linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%);
  color: white;
  border-color: #60a5fa;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
}

.add-btn {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.add-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(139, 92, 246, 0.4);
}

.btn-icon {
  font-size: 1.2rem;
}

/* Services Grid */
.services-grid {
  padding: 40px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 25px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

/* Empty State */
.empty-state {
  padding: 80px 40px;
  text-align: center;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

/* Service Card */
.service-card {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 15px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  border: 2px solid transparent;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
  position: relative;
}

.service-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  border-radius: 3px 3px 0 0;
}

.service-card.online::before {
  background: linear-gradient(90deg, #10b981, #34d399);
}

.service-card.offline::before {
  background: linear-gradient(90deg, #ef4444, #f87171);
}

.service-card.online {
  border-color: rgba(16, 185, 129, 0.3);
}

.service-card.offline {
  border-color: rgba(239, 68, 68, 0.3);
}

.service-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
  border-color: rgba(96, 165, 250, 0.3);
}

.service-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: rgba(15, 23, 42, 0.7);
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.service-title {
  display: flex;
  align-items: center;
  gap: 15px;
}

.service-icon {
  font-size: 2rem;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.service-title h3 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 600;
  color: #e2e8f0;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.status-badge.online {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.status-badge.offline {
  background: rgba(239, 68, 68, 0.2);
  color: #fca5a5;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.service-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
  backdrop-filter: blur(5px);
}

.action-btn.refresh {
  background: rgba(59, 130, 246, 0.2);
  color: #93c5fd;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.action-btn.delete {
  background: rgba(239, 68, 68, 0.2);
  color: #fca5a5;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn:hover {
  transform: scale(1.1);
}

.action-btn.refresh:hover {
  background: rgba(59, 130, 246, 0.3);
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.2);
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.3);
  box-shadow: 0 4px 15px rgba(239, 68, 68, 0.2);
}

.service-details {
  padding: 20px;
}

.detail-row {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.detail-item {
  flex: 1;
}

.detail-label {
  display: block;
  font-size: 0.85rem;
  color: #94a3b8;
  margin-bottom: 6px;
  font-weight: 500;
}

.detail-value {
  font-size: 1rem;
  font-weight: 600;
  color: #e2e8f0;
}

.endpoint {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ip-address {
  color: #60a5fa;
  font-family: 'Courier New', monospace;
}

.port-badge {
  background: rgba(148, 163, 184, 0.1);
  color: #cbd5e1;
  padding: 4px 10px;
  border-radius: 6px;
  font-family: 'Courier New', monospace;
  font-weight: 600;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.last-seen {
  font-style: italic;
}

.last-seen.online {
  color: #34d399;
}

.last-seen.offline {
  color: #f87171;
}

.protocol-badge {
  background: rgba(241, 245, 249, 0.1);
  color: #cbd5e1;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 600;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.service-footer {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}

.uptime-indicator {
  display: flex;
  align-items: center;
  gap: 15px;
}

.uptime-bar {
  flex: 1;
  height: 6px;
  background: rgba(148, 163, 184, 0.1);
  border-radius: 3px;
  overflow: hidden;
}

.uptime-fill {
  height: 100%;
  transition: width 0.5s ease;
}

.uptime-fill.online {
  background: linear-gradient(90deg, #10b981, #34d399);
}

.uptime-fill.offline {
  background: linear-gradient(90deg, #ef4444, #f87171);
}

.uptime-text {
  font-size: 0.9rem;
  color: #94a3b8;
  font-weight: 500;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 20px;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
  opacity: 0.5;
}

.empty-state h3 {
  margin: 0 0 10px;
  color: #e2e8f0;
}

.empty-state p {
  color: #94a3b8;
  margin-bottom: 30px;
}

/* Loading & Error States */
.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.9);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.loading-spinner {
  width: 50px;
  height: 50px;
  border: 5px solid rgba(148, 163, 184, 0.2);
  border-top-color: #60a5fa;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.loading-overlay p {
  margin-top: 20px;
  color: #cbd5e1;
  font-size: 1.1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-banner {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.9) 0%, rgba(220, 38, 38, 0.9) 100%);
  color: white;
  padding: 15px 25px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 15px;
  z-index: 1000;
  box-shadow: 0 4px 20px rgba(239, 68, 68, 0.4);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(254, 202, 202, 0.2);
  min-width: 300px;
}

.error-content {
  display: flex;
  align-items: center;
  gap: 10px;
}

.retry-btn {
  background: rgba(255, 255, 255, 0.2);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.3);
  padding: 6px 12px;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  backdrop-filter: blur(5px);
}

.retry-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(10px);
}

.modal {
  background: rgba(30, 41, 59, 0.95);
  border-radius: 20px;
  width: 500px;
  max-width: 90vw;
  max-height: 90vh;
  overflow-y: auto;
  animation: modalSlide 0.3s ease-out;
  border: 1px solid rgba(148, 163, 184, 0.2);
  backdrop-filter: blur(20px);
}

@keyframes modalSlide {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 25px 30px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: #e2e8f0;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #94a3b8;
  transition: color 0.3s;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #f87171;
  background: rgba(248, 113, 113, 0.1);
}

.modal-content {
  padding: 30px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #cbd5e1;
}

.form-group input {
  width: 100%;
  padding: 12px 15px;
  border: 2px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  font-size: 1rem;
  transition: all 0.3s;
  background: rgba(15, 23, 42, 0.7);
  color: #e2e8f0;
  backdrop-filter: blur(10px);
}

.form-group input:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  background: rgba(15, 23, 42, 0.9);
}

.form-row {
  display: flex;
  gap: 20px;
}

.form-row .form-group {
  flex: 1;
}

.form-note {
  background: rgba(241, 245, 249, 0.1);
  padding: 15px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 25px 0;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

.note-icon {
  font-size: 1.2rem;
  color: #fbbf24;
}

.form-note span {
  color: #cbd5e1;
  font-size: 0.9rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 15px;
  margin-top: 30px;
}

.cancel-btn {
  padding: 12px 24px;
  background: rgba(148, 163, 184, 0.1);
  color: #cbd5e1;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.cancel-btn:hover {
  background: rgba(148, 163, 184, 0.2);
  color: #e2e8f0;
}

.submit-btn {
  padding: 12px 24px;
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(139, 92, 246, 0.4);
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid white;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

/* Confirm Modal */
.confirm-modal {
  background: rgba(30, 41, 59, 0.95);
  border-radius: 20px;
  padding: 40px;
  text-align: center;
  width: 400px;
  max-width: 90vw;
  border: 1px solid rgba(148, 163, 184, 0.2);
  backdrop-filter: blur(20px);
}

.confirm-icon {
  font-size: 3rem;
  margin-bottom: 20px;
  filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.3));
}

.confirm-modal h3 {
  margin: 0 0 15px;
  color: #e2e8f0;
}

.confirm-modal p {
  color: #94a3b8;
  margin-bottom: 30px;
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  gap: 15px;
  justify-content: center;
}

.delete-confirm-btn {
  padding: 12px 24px;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.9) 0%, rgba(220, 38, 38, 0.9) 100%);
  color: white;
  border: none;
  border-radius: 10px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  backdrop-filter: blur(10px);
}

.delete-confirm-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(239, 68, 68, 0.4);
}

/* Scrollbar Styling */
.services-grid::-webkit-scrollbar,
.modal::-webkit-scrollbar,
.confirm-modal::-webkit-scrollbar {
  width: 8px;
}

.services-grid::-webkit-scrollbar-track,
.modal::-webkit-scrollbar-track,
.confirm-modal::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.3);
  border-radius: 10px;
}

.services-grid::-webkit-scrollbar-thumb,
.modal::-webkit-scrollbar-thumb,
.confirm-modal::-webkit-scrollbar-thumb {
  background: rgba(59, 130, 246, 0.5);
  border-radius: 10px;
  border: 2px solid rgba(15, 23, 42, 0.3);
}

.services-grid::-webkit-scrollbar-thumb:hover,
.modal::-webkit-scrollbar-thumb:hover,
.confirm-modal::-webkit-scrollbar-thumb:hover {
  background: rgba(59, 130, 246, 0.7);
}

/* Responsive Design */
@media (max-width: 1024px) {
  .services-grid {
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    padding: 30px;
  }
  
  .header-content,
  .header-stats,
  .action-bar,
  .services-grid,
  .empty-state {
    padding-left: 30px;
    padding-right: 30px;
  }
}

@media (max-width: 768px) {
  .app-wrapper {
    padding: 0;
  }
  
  .app-container {
    border-radius: 0;
  }
  
  .dashboard-header,
  .action-bar,
  .services-grid,
  .empty-state {
    padding: 20px;
  }
  
  .header-stats {
    flex-direction: column;
    gap: 15px;
  }
  
  .action-bar {
    flex-direction: column;
    gap: 20px;
    align-items: stretch;
  }
  
  .search-filter {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-input {
    width: 100%;
  }
  
  .services-grid {
    grid-template-columns: 1fr;
  }
  
  .detail-row {
    flex-direction: column;
    gap: 15px;
  }
  
  .form-row {
    flex-direction: column;
    gap: 15px;
  }
}

@media (max-width: 480px) {
  .dashboard-header {
    padding: 20px 15px;
  }
  
  .header-content h1 {
    font-size: 2rem;
  }
  
  .header-content .subtitle {
    font-size: 1rem;
  }
  
  .action-bar {
    padding: 20px 15px;
  }
  
  .services-grid {
    padding: 20px 15px;
  }
  
  .modal-header {
    padding: 20px;
  }
  
  .modal-content {
    padding: 20px;
  }
  
  .confirm-modal {
    padding: 30px 20px;
  }
  
  .filter-group {
    flex-wrap: wrap;
  }
  
  .filter-btn {
    flex: 1;
    min-width: 100px;
  }
}
</style>
