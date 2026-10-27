import { ref, computed } from 'vue'
import { request } from '../api/client'
import type { User } from '../types'
import { useToast } from './useToast'

const currentUser = ref<User | null>(null)
const token = ref<string | null>(localStorage.getItem('vault_token'))
const isAuthenticating = ref(false)

export function useAuth() {
  const toast = useToast()
  const isAuthenticated = computed(() => !!token.value && !!currentUser.value)

  async function checkSession() {
    if (!token.value) {
      currentUser.value = null
      return
    }
    try {
      const u = await request<User>('/auth/me')
      currentUser.value = u
    } catch {
      logout()
    }
  }

  async function login(ident: string, pass: string) {
    isAuthenticating.value = true
    try {
      const res = await request<{ token: string; user: User }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username: ident, password: pass }),
      })
      token.value = res.token
      currentUser.value = res.user
      localStorage.setItem('vault_token', res.token)
      toast.show(`AUTHENTICATED // Operator: ${res.user.username}`, 'SUCCESS')
      return true
    } catch (err: any) {
      toast.show(err.message || 'Authentication rejected', 'ERROR')
      return false
    } finally {
      isAuthenticating.value = false
    }
  }

  async function register(username: string, email: string, pass: string) {
    isAuthenticating.value = true
    try {
      const res = await request<{ token: string; user: User }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ username, email, password: pass }),
      })
      token.value = res.token
      currentUser.value = res.user
      localStorage.setItem('vault_token', res.token)
      toast.show(`OPERATOR INITIALIZED // ${username}`, 'SUCCESS')
      return true
    } catch (err: any) {
      toast.show(err.message || 'Registration rejected', 'ERROR')
      return false
    } finally {
      isAuthenticating.value = false
    }
  }

  function logout() {
    token.value = null
    currentUser.value = null
    localStorage.removeItem('vault_token')
    toast.show('SESSION TERMINATED', 'WARN')
  }

  return {
    currentUser,
    token,
    isAuthenticated,
    isAuthenticating,
    checkSession,
    login,
    register,
    logout,
  }
}
