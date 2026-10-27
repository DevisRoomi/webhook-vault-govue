import { useQuery, useMutation, useQueryClient } from '@tanstack/vue-query'
import { request } from '../api/client'
import type { Endpoint } from '../types'
import { useToast } from './useToast'

export function useEndpoints() {
  const queryClient = useQueryClient()
  const toast = useToast()

  const endpointsQuery = useQuery({
    queryKey: ['endpoints'],
    queryFn: () => request<Endpoint[]>('/endpoints'),
    staleTime: 15_000,
  })

  const createEndpointMutation = useMutation({
    mutationFn: (newEp: { slug: string; name: string; description: string; target_url: string; secret_token: string }) =>
      request<Endpoint>('/endpoints', {
        method: 'POST',
        body: JSON.stringify(newEp),
      }),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['endpoints'] })
      queryClient.invalidateQueries({ queryKey: ['stats'] })
      toast.show(`INGRESS REGISTERED // slug: ${data.slug}`, 'SUCCESS')
    },
    onError: (err: any) => {
      toast.show(err.message || 'Failed to register ingress', 'ERROR')
    },
  })

  return {
    endpoints: endpointsQuery.data,
    isLoading: endpointsQuery.isLoading,
    isError: endpointsQuery.isError,
    error: endpointsQuery.error,
    refetch: endpointsQuery.refetch,
    createEndpoint: createEndpointMutation.mutateAsync,
    isCreating: createEndpointMutation.isPending,
  }
}
