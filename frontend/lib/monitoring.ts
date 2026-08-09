import apiClient, { apiURL } from "@/lib/axios"

export { apiURL }

export type MonitorStatus = "pending" | "up" | "down"

export type User = {
  id: string
  firstName: string
  lastName: string
  email: string
  createdAt: string
}

export type Monitor = {
  id: string
  userId?: string
  name: string
  url: string
  intervalSeconds?: number
  nextCheckAt?: string
  lastCheckedAt?: string
  lastStatus: MonitorStatus
  lastResponseMs?: number
  createdAt?: string
  updatedAt?: string
}

export type CreateMonitorInput = {
  name: string
  url: string
  intervalSeconds: number
}

export type UpdateMonitorInput = {
  name?: string
  url?: string
  intervalSeconds?: number
  paused?: boolean
}

export type CheckResult = {
  monitorId: string
  status: MonitorStatus
  statusCode: number
  responseMs: number
  errorMessage?: string
  checkedAt: string
}

export type AuthResponse = {
  user?: User
  token: string
}

export async function registerUser(input: {
  firstName: string
  lastName: string
  email: string
  password: string
}): Promise<AuthResponse> {
  const { data } = await apiClient.post<AuthResponse>("/api/v1/auth/register", input)
  return data
}

export async function loginUser(input: {
  email: string
  password: string
}): Promise<{ token: string }> {
  const { data } = await apiClient.post<{ token: string }>("/api/v1/auth/login", input)
  return data
}

export async function getMe(): Promise<User> {
  const { data } = await apiClient.get<User>("/api/v1/auth/me")
  return data
}

export async function getMonitors(token?: string): Promise<Monitor[]> {
  const headers = token ? { Authorization: `Bearer ${token}` } : undefined
  const { data } = await apiClient.get<Monitor[]>("/api/v1/monitors/", { headers })
  return data
}

export async function getMonitorById(id: string): Promise<Monitor> {
  const { data } = await apiClient.get<Monitor>(`/api/v1/monitors/${id}`)
  return data
}

export async function getMonitorHistory(id: string, limit = 50): Promise<CheckResult[]> {
  const { data } = await apiClient.get<CheckResult[]>(
    `/api/v1/monitors/${id}/history`,
    { params: { limit } },
  )
  return data
}

export async function createMonitor(input: CreateMonitorInput): Promise<Monitor> {
  const { data } = await apiClient.post<Monitor>("/api/v1/monitors/", input)
  return data
}

export async function updateMonitor(id: string, input: UpdateMonitorInput): Promise<Monitor> {
  const { data } = await apiClient.patch<Monitor>(`/api/v1/monitors/${id}`, input)
  return data
}

export async function deleteMonitor(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/monitors/${id}`)
}

export async function getPublicMonitors(): Promise<Monitor[]> {
  const { data } = await apiClient.get<Monitor[]>("/api/status")
  return data
}
