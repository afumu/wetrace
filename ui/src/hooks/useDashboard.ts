import { useQuery } from '@tanstack/react-query'
import { dashboardApi } from '@/api/dashboard'

export const useDashboard = () => {
  return useQuery({
    queryKey: ['dashboard'],
    queryFn: dashboardApi.getDashboardData
  })
}
