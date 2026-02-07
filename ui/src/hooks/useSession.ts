import { useQuery } from '@tanstack/react-query'
import { sessionApi } from '@/api/session'
import type { SessionParams } from '@/types'

export const useSessions = (params?: SessionParams) => {
  return useQuery({
    queryKey: ['sessions', params],
    queryFn: async () => {
      const { items } = await sessionApi.getSessions({ 
        ...params, 
        offset: 0, 
        limit: 10000
      })
      return items
    }
  })
}