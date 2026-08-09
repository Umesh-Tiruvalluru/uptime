"use client"

import { useEffect, useState, useMemo, useCallback } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import {
  apiURL,
  getMonitors,
  getPublicMonitors,
  type CheckResult,
  type Monitor,
} from "@/lib/monitoring"
import { useAuth } from "@/context/auth-context"
import { MonitorDialog } from "@/components/monitor-dialog"
import { DeleteDialog } from "@/components/delete-dialog"

type MonitorListProps = {
  publicView?: boolean
  refreshTrigger?: number
}

type FilterStatus = "all" | "up" | "down" | "pending"

export function MonitorList({ publicView = false, refreshTrigger = 0 }: MonitorListProps) {
  const router = useRouter()
  const { token } = useAuth()
  const [monitors, setMonitors] = useState<Monitor[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [connected, setConnected] = useState(false)

  // Filtering & Search
  const [search, setSearch] = useState("")
  const [statusFilter, setStatusFilter] = useState<FilterStatus>("all")

  // Modals state
  const [editingMonitor, setEditingMonitor] = useState<Monitor | null>(null)
  const [deletingMonitor, setDeletingMonitor] = useState<Monitor | null>(null)

  const loadMonitors = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      if (publicView) {
        setMonitors(await getPublicMonitors())
      } else {
        const authToken =
          token ??
          window.localStorage.getItem("monitoring_token") ??
          window.localStorage.getItem("token")

        if (!authToken) {
          throw new Error("Please sign in to view and manage your monitors.")
        }
        setMonitors(await getMonitors(authToken))
      }
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to load monitors.")
    } finally {
      setLoading(false)
    }
  }, [publicView, token])

  useEffect(() => {
    void loadMonitors()
  }, [loadMonitors, refreshTrigger])

  // Real-time SSE Connection
  useEffect(() => {
    const stream = new EventSource(`${apiURL}/api/events`)
    stream.onopen = () => setConnected(true)
    stream.onerror = () => setConnected(false)

    const onCheckCompleted = (event: Event) => {
      try {
        const result = JSON.parse((event as MessageEvent<string>).data) as CheckResult
        setMonitors((current) =>
          current.map((monitor) =>
            monitor.id === result.monitorId
              ? {
                  ...monitor,
                  lastStatus: result.status,
                  lastCheckedAt: result.checkedAt,
                  lastResponseMs: result.responseMs,
                }
              : monitor,
          ),
        )
      } catch {
        // Ignore malformed events
      }
    }

    stream.addEventListener("check.completed", onCheckCompleted)
    return () => {
      stream.removeEventListener("check.completed", onCheckCompleted)
      stream.close()
    }
  }, [])

  // Filtered monitors list
  const filteredMonitors = useMemo(() => {
    return monitors.filter((m) => {
      const matchesSearch =
        m.name.toLowerCase().includes(search.toLowerCase()) ||
        m.url.toLowerCase().includes(search.toLowerCase())
      const matchesStatus =
        statusFilter === "all" || m.lastStatus === statusFilter
      return matchesSearch && matchesStatus
    })
  }, [monitors, search, statusFilter])

  // Summary Metrics
  const stats = useMemo(() => {
    const total = monitors.length
    const up = monitors.filter((m) => m.lastStatus === "up").length
    const down = monitors.filter((m) => m.lastStatus === "down").length

    const validLatencies = monitors
      .map((m) => m.lastResponseMs)
      .filter((ms): ms is number => ms !== undefined && ms !== null)

    const avgLatency =
      validLatencies.length > 0
        ? Math.round(validLatencies.reduce((a, b) => a + b, 0) / validLatencies.length)
        : 0

    return { total, up, down, avgLatency }
  }, [monitors])

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="h-10 w-full animate-pulse rounded-xl bg-muted" />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-2">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="h-36 rounded-2xl border border-border/60 bg-card p-5 animate-pulse" />
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-2xl border border-destructive/30 bg-destructive/10 p-6 text-center">
        <p className="text-sm font-medium text-destructive mb-3">{error}</p>
        {!publicView && (
          <button
            onClick={() => void loadMonitors()}
            className="rounded-xl border border-destructive/40 bg-card px-4 py-1.5 text-xs font-semibold text-destructive hover:bg-destructive/10 transition-colors"
          >
            Retry Loading
          </button>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Stats Summary Bar */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        <div className="rounded-2xl border border-border/60 bg-card p-4 shadow-2xs">
          <span className="text-xs font-medium text-muted-foreground">Total Monitors</span>
          <p className="mt-1 text-2xl font-bold tracking-tight">{stats.total}</p>
        </div>
        <div className="rounded-2xl border border-emerald-500/20 bg-emerald-500/5 p-4 shadow-2xs">
          <span className="text-xs font-medium text-emerald-600 dark:text-emerald-400">Operational</span>
          <p className="mt-1 text-2xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400">
            {stats.up}
          </p>
        </div>
        <div className="rounded-2xl border border-rose-500/20 bg-rose-500/5 p-4 shadow-2xs">
          <span className="text-xs font-medium text-rose-600 dark:text-rose-400">Down / Issues</span>
          <p className="mt-1 text-2xl font-bold tracking-tight text-rose-600 dark:text-rose-400">
            {stats.down}
          </p>
        </div>
        <div className="rounded-2xl border border-border/60 bg-card p-4 shadow-2xs">
          <span className="text-xs font-medium text-muted-foreground">Avg Response</span>
          <p className="mt-1 text-2xl font-bold tracking-tight">
            {stats.avgLatency ? `${stats.avgLatency} ms` : "—"}
          </p>
        </div>
      </div>

      {/* Control Bar: Search & Status Filters */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
        <div className="relative flex-1 max-w-sm">
          <svg className="absolute left-3 top-2.5 size-4 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by name or URL..."
            className="w-full rounded-xl border border-input bg-card pl-9 pr-3 py-2 text-xs focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1 rounded-xl border border-border bg-card p-1">
            {(["all", "up", "down", "pending"] as const).map((st) => (
              <button
                key={st}
                onClick={() => setStatusFilter(st)}
                className={`rounded-lg px-2.5 py-1 text-xs font-semibold capitalize transition-all ${
                  statusFilter === st
                    ? "bg-primary text-primary-foreground shadow-2xs"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {st}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-1.5 pl-2 text-xs text-muted-foreground" title="SSE Connection Status">
            <span className={`size-2.5 rounded-full transition-colors ${connected ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]" : "bg-amber-500 animate-pulse"}`} />
            <span className="hidden sm:inline">{connected ? "Live Stream" : "Connecting"}</span>
          </div>
        </div>
      </div>

      {/* Grid of Monitors */}
      {filteredMonitors.length === 0 ? (
        <div className="rounded-2xl border border-dashed border-border/80 bg-card/50 p-12 text-center">
          <svg className="mx-auto size-10 text-muted-foreground/60 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
          </svg>
          <p className="text-sm font-medium">No monitors found</p>
          <p className="text-xs text-muted-foreground mt-1">
            {monitors.length === 0
              ? "Get started by adding your first endpoint monitor."
              : "Try adjusting your search query or status filter."}
          </p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-2">
          {filteredMonitors.map((m) => (
            <MonitorCard
              key={m.id}
              monitor={m}
              publicView={publicView}
              onEdit={() => setEditingMonitor(m)}
              onDelete={() => setDeletingMonitor(m)}
            />
          ))}
        </div>
      )}

      {/* Modals */}
      {!publicView && (
        <>
          <MonitorDialog
            isOpen={Boolean(editingMonitor)}
            monitorToEdit={editingMonitor}
            onClose={() => setEditingMonitor(null)}
            onSuccess={() => {
              setEditingMonitor(null)
              void loadMonitors()
            }}
          />
          <DeleteDialog
            isOpen={Boolean(deletingMonitor)}
            monitor={deletingMonitor}
            onClose={() => setDeletingMonitor(null)}
            onSuccess={() => {
              setDeletingMonitor(null)
              void loadMonitors()
            }}
          />
        </>
      )}
    </div>
  )
}

function MonitorCard({
  monitor,
  publicView,
  onEdit,
  onDelete,
}: {
  monitor: Monitor
  publicView: boolean
  onEdit: () => void
  onDelete: () => void
}) {
  const isUp = monitor.lastStatus === "up"
  const isPending = monitor.lastStatus === "pending"
  const label = isPending ? "Checking" : isUp ? "Operational" : "Down"

  return (
    <article className="group relative rounded-2xl border border-border/80 bg-card p-5 shadow-xs transition-all hover:border-border hover:shadow-md space-y-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <Link
            href={`/monitors/${monitor.id}`}
            className="text-left font-bold text-base hover:text-primary transition-colors truncate block max-w-full"
          >
            {monitor.name}
          </Link>
          <a
            href={monitor.url}
            target="_blank"
            rel="noreferrer"
            className="truncate text-xs text-muted-foreground hover:underline font-mono block mt-0.5"
          >
            {monitor.url}
          </a>
        </div>

        <span
          className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-semibold transition-colors ${
            isPending
              ? "bg-amber-500/15 text-amber-700 dark:text-amber-300"
              : isUp
              ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
              : "bg-destructive/15 text-destructive"
          }`}
        >
          {label}
        </span>
      </div>

      <dl className="grid grid-cols-2 gap-4 border-t border-border/40 pt-3 text-xs">
        <div>
          <dt className="text-muted-foreground font-medium">Response Latency</dt>
          <dd className="mt-1 font-semibold text-foreground">
            {monitor.lastResponseMs === undefined ? "—" : `${monitor.lastResponseMs} ms`}
          </dd>
        </div>
        <div>
          <dt className="text-muted-foreground font-medium">Last Checked</dt>
          <dd className="mt-1 font-semibold text-foreground">
            {formatDate(monitor.lastCheckedAt)}
          </dd>
        </div>
      </dl>

      <div className="flex items-center justify-between border-t border-border/40 pt-3">
        <Link
          href={`/monitors/${monitor.id}`}
          className="text-xs font-semibold text-primary hover:underline flex items-center gap-1"
        >
          View Insights Page
          <svg className="size-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
          </svg>
        </Link>

        {!publicView && (
          <div className="flex items-center gap-1">
            <button
              onClick={onEdit}
              title="Edit Monitor"
              className="rounded-lg p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
            >
              <svg className="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
              </svg>
            </button>

            <button
              onClick={onDelete}
              title="Delete Monitor"
              className="rounded-lg p-1.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive transition-colors"
            >
              <svg className="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
              </svg>
            </button>
          </div>
        )}
      </div>
    </article>
  )
}

function formatDate(value?: string) {
  if (!value) return "Not yet checked"
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? "Not yet checked" : date.toLocaleString()
}
