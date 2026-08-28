<template>
  <div class="auth-page">
    <div class="auth-card">
      <div class="auth-logo">
        <img :src="logoURL" alt="TailPass" />
      </div>
      <h1>Sign in to TailPass</h1>
      <p class="subtitle">Enter your credentials to manage your tunnels</p>

      <form @submit.prevent="submit">
        <div class="form-group">
          <label for="username">Username</label>
          <input id="username" v-model="username" autocomplete="username" required autofocus />
        </div>
        <div class="form-group">
          <label for="password">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            required
          />
        </div>

        <p v-if="error" class="error-text">{{ error }}</p>

        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? 'Signing in…' : 'Sign In' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import logoURL from '../assets/logo.png'

const username = ref('')
const password = ref('')
const error = ref(null)
const loading = ref(false)

const auth = useAuthStore()
const router = useRouter()

async function submit() {
  error.value = null
  loading.value = true
  try {
    await auth.login(username.value, password.value)
    router.push('/')
  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  padding: 20px;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: rgba(30, 41, 59, 0.9);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 18px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(10px);
}

.auth-logo {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.auth-logo img {
  width: 56px;
}

h1 {
  color: #e2e8f0;
  text-align: center;
  font-size: 1.6rem;
  margin: 0 0 6px;
}

.subtitle {
  text-align: center;
  color: #94a3b8;
  margin: 0 0 28px;
  font-size: 0.95rem;
}

.form-group {
  margin-bottom: 18px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #cbd5e1;
  font-size: 0.9rem;
}

input {
  width: 100%;
  padding: 12px 15px;
  border: 2px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  font-size: 1rem;
  background: rgba(15, 23, 42, 0.7);
  color: #e2e8f0;
  transition: all 0.3s;
}

input:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 3px rgba(96, 165, 250, 0.1);
  background: rgba(15, 23, 42, 0.9);
}

.error-text {
  color: #f87171;
  font-size: 0.9rem;
  margin: -6px 0 16px;
}

.submit-btn {
  width: 100%;
  padding: 13px;
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
  border-radius: 10px;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.3s;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(139, 92, 246, 0.4);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>
