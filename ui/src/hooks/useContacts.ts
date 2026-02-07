import { useQuery } from '@tanstack/react-query'
import { contactApi } from '@/api/contact'
import type { Contact, ContactParams } from '@/types'

export const useContacts = (params?: ContactParams) => {
  return useQuery({
    queryKey: ['contacts', params],
    queryFn: async () => {
      const contacts = await contactApi.getContacts(params)
      return contacts
    }
  })
}

// Helper to group contacts by initial
export const groupContacts = (contacts: Contact[]) => {
  const groups: Record<string, Contact[]> = {}
  
  contacts.forEach(contact => {
    // Basic grouping logic - can be improved with pinyin library
    let initial = (contact.remark || contact.nickname || contact.wxid).charAt(0).toUpperCase()
    if (!/[A-Z]/.test(initial)) initial = '#'
    
    if (!groups[initial]) groups[initial] = []
    groups[initial].push(contact)
  })
  
  return Object.keys(groups).sort().reduce((acc, key) => {
    if (key === '#') return acc // Move # to end
    acc[key] = groups[key]
    return acc
  }, { ...groups, '#': groups['#'] } as Record<string, Contact[]>)
}
