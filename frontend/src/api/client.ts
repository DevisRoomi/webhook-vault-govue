const API_BASE = '/api'

export async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('vault_token')
  const headers = new Headers(options.headers || {})

  if (!headers.has('Content-Type') && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  })

  const json = await res.json().catch(() => null)

  if (!res.ok) {
    const errorMsg = json?.error || json?.message || `HTTP ${res.status}: ${res.statusText}`
    throw new Error(errorMsg)
  }

  return (json?.data !== undefined ? json.data : json) as T
}
