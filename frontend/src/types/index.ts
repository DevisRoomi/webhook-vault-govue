export interface User {
  id: number
  username: string
  email: string
  role: string
  created_at: string
}

export interface Endpoint {
  id: number
  slug: string
  name: string
  description: string
  target_url: string
  secret_token: string
  is_active: boolean
  total_received: number
  total_forwarded: number
  created_at: string
}

export interface WebhookEventSummary {
  id: number
  endpoint_id: number
  request_id: string
  method: string
  path: string
  headers: string
  query_params: string
  body: string
  content_type: string
  source_ip: string
  byte_size: number
  status: 'RECEIVED' | 'FORWARDED' | 'FAILED' | 'RETRIED'
  created_at: string
  endpoint_slug: string
  endpoint_name: string
  target_url: string
  latest_http_status?: number
  latest_latency_ms?: number
}

export interface DeliveryAttempt {
  id: number
  event_id: number
  target_url: string
  attempt_num: number
  http_status: number
  response_headers: string
  response_body: string
  latency_ms: number
  error_message: string
  dispatched_at: string
}

export interface EventDetail extends WebhookEventSummary {
  attempts: DeliveryAttempt[]
}

export interface TelemetryStats {
  total_events: number
  delivered_count: number
  failed_count: number
  pending_count: number
  active_endpoints: number
  uptime_seconds: number
}

export interface FilterState {
  endpointId: string
  status: string
  search: string
}
