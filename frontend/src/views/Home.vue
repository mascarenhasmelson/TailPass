
<template>


  <main class="home-page">
    <div class="Dashboard-info">

      <div class="home-card">
        <b>Public IP:</b> {{ Public_IP }}
      </div>

      <div class="home-card">
        <b>ISP Info:</b> {{ ISP_Info }}
      </div>

      <div class="home-card">
        <b>Internet Connection Status:</b> {{ Internet_Status }}
      </div>

    </div>
  </main>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue';

const API_URL = import.meta.env.VITE_API_URL;


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

</style>
