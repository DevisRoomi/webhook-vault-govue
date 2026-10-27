import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { request } from '../api/client'
import type { DeliveryAttempt } from '../types'
import { useToast } from './useToast'

export function useDispatchRetry() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const retryMutation = useMutation({
    mutationFn: ({ eventId, overrideTargetUrl }: { eventId: number; overrideTargetUrl?: string }) =>
      request<DeliveryAttempt>(`/events/${eventId}/retry`, {
        method: 'POST',
        body: JSON.stringify({ override_target_url: overrideTargetUrl || '' }),
      }),
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['events'] })
      queryClient.invalidateQueries({ queryKey: ['event-detail', variables.eventId] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })

      if (data.http_status >= 200 && data.http_status < 300) {
        toast.show(`DISPATCH CONFIRMED // Attempt #${data.attempt_num}: HTTP ${data.http_status} [${data.latency_ms}ms]`, 'SUCCESS')
      } else {
        toast.show(`DISPATCH FAILED // Attempt #${data.attempt_num}: HTTP ${data.http_status || 'ERR'} - ${data.error_message}`, 'ERROR')
      }
    },
    onError: (err: any) => {
      toast.show(err.message || 'Retry dispatch failed', 'ERROR')
    },
  })

  const simulateMutation = useMutation({
    mutationFn: (req: { event_name: string; simulate: string }) =>
      request('/test/simulate', {
        method: 'POST',
        body: JSON.stringify(req),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
      toast.show('SYNTHETIC EVENT COMMITTED // Ingress stream updated', 'INFO')
    },
    onError: (err: any) => {
      toast.show(err.message || 'Simulation dispatch rejected', 'ERROR')
    },
  })

  return {
    retryDispatch: retryMutation.mutateAsync,
    isRetrying: retryMutation.isPending,
    simulateTest: simulateMutation.mutateAsync,
    isSimulating: simulateMutation.isPending,
  }
}
