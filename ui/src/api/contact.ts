import { request, getApiBaseUrl } from "@/lib/request"
import { ContactType } from "@/types"
import type { Contact, ContactParams } from "@/types"

interface BackendContact {
  userName: string
  alias: string
  remark: string
  nickName: string
  isFriend: boolean
}

function getAvatarUrl(username?: string): string {
  if (!username) return ''
  return `/avatar/${username}`
}

function transformContact(backendContact: BackendContact): Contact {
  let type: ContactType
  if (backendContact.userName.endsWith('@chatroom')) {
    type = ContactType.Chatroom
  } else if (backendContact.userName.startsWith('gh_')) {
    type = ContactType.Official
  } else {
    type = ContactType.Friend
  }

  const avatar = getAvatarUrl(backendContact.userName)

  return {
    wxid: backendContact.userName,
    nickname: backendContact.nickName || backendContact.userName,
    remark: backendContact.remark || '',
    alias: backendContact.alias || '',
    avatar,
    type,
    isStarred: false,
    isPinned: false,
    isMinimized: false,
    bigHeadImgUrl: '',
    smallHeadImgUrl: '',
    headImgMd5: '',
  }
}

export interface NeedContactItem {
  userName: string
  nickName: string
  remark: string
  lastContactTime: number
  daysSinceContact: number
}

export const contactApi = {
  getContacts: async (params?: ContactParams): Promise<Contact[]> => {
    const response = await request.get<BackendContact[]>('/api/v1/contacts', params)
    
    if (Array.isArray(response)) {
      return response.map(transformContact)
    }
    
    return []
  },

  getContactDetail: async (wxid: string): Promise<Contact> => {
    // Note: Detail endpoint usually remains singular or appends ID to plural
    // Assuming /api/v1/contacts/:id based on REST conventions, but doc doesn't specify detail endpoint.
    // Keeping singular /api/v1/contact/:id as fallback or guessing /api/v1/contacts/:id?
    // The doc only shows list endpoints. Let's assume standard REST: /api/v1/contacts/:id
    const response = await request.get<BackendContact>(`/api/v1/contacts/${encodeURIComponent(wxid)}`)
    return transformContact(response)
  },

  searchContacts: async (keyword: string): Promise<Contact[]> => {
    return contactApi.getContacts({ keyword })
  },

  getChatroomMembers: async (chatroomId: string): Promise<Contact[]> => {
    const chatroom = await contactApi.getContactDetail(chatroomId)
    if (!chatroom.memberList) {
      return []
    }

    const memberPromises = chatroom.memberList.map(wxid =>
      contactApi.getContactDetail(wxid).catch(() => null)
    )
    const members = await Promise.all(memberPromises)

    return members.filter((m): m is Contact => m !== null)
  },

  getDisplayName: (contact: Contact): string => {
    return contact.remark || contact.nickname || contact.alias || contact.wxid
  },

  exportContacts: (format: 'csv' | 'xlsx' = 'csv', keyword?: string): string => {
    const baseURL = getApiBaseUrl()
    const params = new URLSearchParams({ format })
    if (keyword) params.set('keyword', keyword)
    return `${baseURL}/api/v1/contacts/export?${params.toString()}`
  },

  getNeedContactList: async (days: number = 7): Promise<NeedContactItem[]> => {
    const response = await request.get<NeedContactItem[]>('/api/v1/contacts/need-contact', { days })
    if (Array.isArray(response)) {
      return response
    }
    return []
  },
}

export { getAvatarUrl }
