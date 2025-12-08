
<template>


  <main class="home-page">
    <div class="Dashboard-info">

      <div class="home-card">Public IP 
       <span class="highlight-ip">  {{ Public_IP }}</span>
      </div>

      <div class="home-card">ISP Info
       <span class="highlight-ip"> {{ ISP_Info }}</span>
      </div>

 <div
  class="home-card"
  :class="Internet_Status === 'Online' ? 'online' : 'offline'"
>
  <b>Internet Connection Status</b> {{ Internet_Status }}
</div>

    </div>
  </main>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue';

// const API_URL = import.meta.env.VITE_API_URL;
const API_URL = "http://192.168.20.17:8082";
export default {
  data() {
    return {
      Public_IP: "Loading...",
      ISP_Info: "Loading...",
      Internet_Status: "Checking...",
      intervalId: null, 
    };
  },

  methods: {
    async fetchISPInfo() {
      try {
        const response = await fetch(`${API_URL}/services/isp`);
        
        if (!response.ok) {
          throw new Error("Failed to fetch ISP info");
        }

        const data = await response.json();

      
        this.Public_IP = data.ip || "Unknown";
        this.ISP_Info = data.org || "Unknown";
        this.Internet_Status = "Online";
      } catch (error) {
        console.error("Error fetching ISP info:", error);
        this.Internet_Status = "Offline";
      }
    },
  },

   mounted() {
    this.fetchISPInfo();
    this.intervalId = setInterval(() => {
      this.fetchISPInfo();
    }, 10000);
  },
  beforeUnmount() {
    clearInterval(this.intervalId);
  }
};
</script>

<style scoped>

body.dark {
  --bg: #0b1220;
  --card-bg: #0f1724;
  --text: #e6eef8;
  --muted: #9ca3af;
  --glow: rgba(59, 130, 246, 0.35); /* Blue glow */
  --glow-strong: rgba(59, 130, 246, 0.55);
}

.home-page {
  background: var(--bg);
  min-height: 100vh;
  padding: 40px;
}

.Dashboard-info {
  display: flex;
  justify-content: flex-start;
  align-items: flex-start;
  gap: 30px;
  margin-top: 20px;
}


.home-card {
  background: var(--card-bg);
  color: var(--text);
  width: 320px;
  height: 180px;
  border-radius: 16px;
  padding: 24px;

  box-shadow:
    0 0 14px rgba(0, 0, 0, 0.6),
    0 0 12px var(--glow);


  transition: all 0.25s ease-in-out;

  display: flex;
  flex-direction: column;
  justify-content: center;
}
.home-card:hover {
  transform: translateY(-6px);
  box-shadow:
    0 0 20px rgba(0, 0, 0, 0.8),
    0 0 22px var(--glow-strong),
    0 0 40px var(--glow);
}

.home-card b {
  font-size: 20px;
  font-weight: 600;
}
.home-card.online {
  border: 2px solid #4ade80;            
  box-shadow:
    0 0 14px rgba(0, 0, 0, 0.6),
    0 0 12px rgba(34, 197, 94, 0.6);     
}
.home-card.offline {
  border: 2px solid #f87171;            
  box-shadow:
    0 0 14px rgba(0, 0, 0, 0.6),
    0 0 12px rgba(248, 113, 113, 0.6);  
}
.home-card {
  border: 2px solid #2813dd;            
  box-shadow:
    0 0 14px rgba(0, 0, 0, 0.6),
    0 0 12px rgba(9, 20, 177, 0.6);     
}
body.dark .highlight-ip {
  color: #60a5fa;
  background: rgba(96, 165, 250, 0.15);
}
.highlight-ip {
  font-weight: 600;
  color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
}
</style>
