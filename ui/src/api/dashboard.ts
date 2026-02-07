import { request } from "@/lib/request"

export interface DashboardData {
  overview: {
    user: string
    dbStats: {
      db_size_mb: number
      dir_size_mb: number
    }
    msgStats: {
      total_msgs: number
      sent_msgs: number
      received_msgs: number
      unique_msg_types: number
    }
    msgTypes: Record<string, number>
    groups: Array<{
      ChatRoomName: string
      NickName: string
      member_count: number
      message_count: number
    }>
    timeline: {
      earliest_msg_time: number
      latest_msg_time: number
      duration_days: number
    }
  }
  visualization: {
    defaults: {
      selectedGroupIndex: number
    }
    groupAnalysis: {
      title: string
      overview: {
        total_groups: number
        active_groups: number
        today_messages: number
        weekly_avg: number
        most_active_hour: string
      }
      content_analysis: {
        text_messages: number
        images: number
        voice_messages: number
        files: number
        links: number
        others: number
      }
      group_list: Array<{
        name: string
        members: number
        messages: number
        active: boolean
      }>
    }
    dataTypeAnalysis: {
      title: string
      content_types: Record<string, {
        count: number
        percentage: number
        size_mb?: number
        trend?: string
      }>
      source_channels: Record<string, {
        count: number
        percentage: number
      }>
      processing_status: {
        processed: number
        processing: number
        pending: number
      }
      quality_metrics: {
        data_integrity: number
        classification_accuracy: number
        duplicate_rate: number
        error_rate: number
      }
      pieGradient: string
    }
    relationshipNetwork: {
      nodes: Array<{
        name: string
        type: string
        messages: number
        avatar: string
      }>
    }
  }
}

export const dashboardApi = {
  getDashboardData: async () => {
    return request.get<DashboardData>('/api/v1/dashboard')
  }
}
