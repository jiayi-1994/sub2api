/**
 * Public (unauthenticated) Channel Monitor API.
 * Mirrors api/channelMonitor.ts but targets /public/monitoring endpoints
 * which do not require auth and omit sensitive fields (e.g. group_name).
 */

import { apiClient } from './client'
import type { Provider, MonitorStatus } from './admin/channelMonitor'
import type { MonitorTimelinePoint, UserMonitorExtraModel, UserMonitorModelDetail } from './channelMonitor'

export type { Provider, MonitorStatus } from './admin/channelMonitor'
export type { MonitorTimelinePoint, UserMonitorExtraModel, UserMonitorModelDetail }

export interface PublicMonitorView {
  id: number
  name: string
  provider: Provider
  primary_model: string
  primary_status: MonitorStatus
  primary_latency_ms: number | null
  primary_ping_latency_ms: number | null
  availability_7d: number
  extra_models: UserMonitorExtraModel[]
  timeline: MonitorTimelinePoint[]
}

export interface PublicMonitorListResponse {
  items: PublicMonitorView[]
}

export interface PublicMonitorDetail {
  id: number
  name: string
  provider: Provider
  models: UserMonitorModelDetail[]
}

export async function list(options?: { signal?: AbortSignal }): Promise<PublicMonitorListResponse> {
  const { data } = await apiClient.get<PublicMonitorListResponse>('/public/monitoring', {
    signal: options?.signal,
  })
  return data
}

export async function status(id: number): Promise<PublicMonitorDetail> {
  const { data } = await apiClient.get<PublicMonitorDetail>(`/public/monitoring/${id}/status`)
  return data
}

export const channelMonitorPublicAPI = { list, status }
export default channelMonitorPublicAPI
