import { useQuery } from '@tanstack/react-query'
import { chatlogApi } from '@/api/chatlog'

export const useMessages = (talker: string | null) => {
  return useQuery({
    queryKey: ['messages', talker],
    queryFn: async () => {
      if (!talker) return []
      
      // Load a large number of messages once, as per user requirement
      // Using a very large limit and time range that covers "all" if possible, 
      // or just trust getSessionMessages with large limit.
      // Assuming 'getSessionMessages' with time=undefined and large limit works.
      const items = await chatlogApi.getSessionMessages(talker, "2000-01-01~2099-12-31", 1000000, 0)
      
      // We want oldest at the beginning, newest at the end for the message list.
      // If the API returns newest first (DESC), we reverse it.
      // Most chat APIs return ASC or allow specifying. Let's check transformMessage.
      // Usually, we want [oldest, ..., newest] for rendering.
      return items.sort((a, b) => a.createTime - b.createTime)
    },
    enabled: !!talker,
  })
}