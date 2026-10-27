import type { Ref } from 'vue'
import { ref } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { request } from '../api/client'
import type { WebhookEventSummary, EventDetail, TelemetryStats, FilterState } from '../types'

export function useWebhooks(filter: Ref<FilterState>) {
  const selectedEventId = ref<number | null>(null)

  // Events list query
  const eventsQuery = useQuery({
    queryKey: ['events', filter],
    queryFn: () => {
      const p = new URLSearchParams()
      if (filter.value.endpointId && filter.value.endpointId !== 'ALL') {
        p.set('endpoint_id', filter.value.endpointId)
      }
      if (filter.value.status && filter.value.status !== 'ALL') {
        p.set('status', filter.value.status)
      }
      if (filter.value.search) {
        p.set('search', filter.value.search)
      }
      const qs = p.toString() ? `?${p.toString()}` : ''
      return request<WebhookEventSummary[]>(`/events${qs}`)
    },
    staleTime: 5_000,
    refetchInterval: 5_000, // 5s auto polling for real-time ingress monitor
  })

  // Deep event inspector query
  const eventDetailQuery = useQuery({
    queryKey: ['event-detail', selectedEventId],
    queryFn: () => {
      if (!selectedEventId.value) return null
      return request<EventDetail>(`/events/${selectedEventId.value}`)
    },
    enabled: () => !!selectedEventId.value,
  })

  // Telemetry stats
  const statsQuery = useQuery({
    queryKey: ['stats'],
    queryFn: () => request<TelemetryStats>('/stats'),
    staleTime: 5_000,
    refetchInterval: 5_000,
  })

  function selectEvent(id: number | null) {
    selectedEventId.value = id
  }

  return {
    events: eventsQuery.data,
    isLoading: eventsQuery.isLoading,
    isFetching: eventsQuery.isFetching,
    isError: eventsQuery.isError,
    error: eventsQuery.error,
    refetchEvents: eventsQuery.refetch,

    selectedEventId,
    selectedEvent: eventDetailQuery.data,
    isLoadingDetail: eventDetailQuery.isLoading,
    refetchDetail: eventDetailQuery.refetch,
    selectEvent,

    stats: statsQuery.data,
    refetchStats: statsQuery.refetch,
  }
}
