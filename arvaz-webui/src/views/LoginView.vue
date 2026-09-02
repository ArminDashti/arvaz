<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { setToken } from '@/lib/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const router = useRouter()
const username = ref('armin')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const resp = await api.login(username.value, password.value)
    setToken(resp.token)
    await router.push('/')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <Card class="login-card">
      <CardHeader>
        <CardTitle>Arvaz</CardTitle>
        <p class="text-sm text-muted-foreground">Sign in to the host dashboard</p>
      </CardHeader>
      <CardContent>
        <form class="space-y-4" @submit.prevent="submit">
          <label class="block space-y-1">
            <span class="text-sm font-medium">Username</span>
            <input v-model="username" class="field-input" autocomplete="username" required />
          </label>
          <label class="block space-y-1">
            <span class="text-sm font-medium">Password</span>
            <input v-model="password" class="field-input" type="password" autocomplete="current-password" required />
          </label>
          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
          <button type="submit" class="field-button" :disabled="loading">
            {{ loading ? 'Signing in…' : 'Sign in' }}
          </button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
