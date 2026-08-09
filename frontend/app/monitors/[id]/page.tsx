"use client"

import { useEffect, useState, useCallback } from "react"
import { useParams, useRouter } from "next/navigation"
import Link from "next/link"
import {
  apiURL,
  getMonitorById,
  getMonitorHistory,
  type CheckResult,
  type Monitor,
} from "@/lib/monitoring"
import { LatencyGraph, type LatencyPoint } from "@/components/latency-graph"
import { MonitorDialog } from "@/components/monitor-dialog"
import { DeleteDialog } from "@/components/delete-dialog"
import { useAuth } from "@/context/auth-context"

export default function MonitorDetailsPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const monitorId = params.id as string

  const [monitor, setMonitor] = useState<Monitor | null>(null)
  const [history, setHistory] = useState<LatencyPoint[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [connected, setConnected] = useState(false)

  // Dialogs state
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)

  const fetchMonitor = useCallback(async () => {
    if (!monitorId) return
    setLoading(true)
    setError(null)
    try {
      const data = await getMonitorById(monitorId)
      setMonitor(data)

      try {
        const historyFromDb = await getMonitorHistory(monitorId, 50)
        setHistory(
          historyFromDb.map((r) => ({
            ms: r.responseMs,
            status: r.status,
            timestamp: r.checkedAt,
          })),
        )
      } catch {
        if (data.lastResponseMs !== undefined) {
          setHistory([
            {
              ms: data.lastResponseMs,
              status: data.lastStatus,
              timestamp: data.lastCheckedAt,
            },
          ])
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load monitor details.")
    } finally {
      setLoading(false)
    }
  }, [monitorId])

  useEffect(() => {
    void fetchMonitor()
  }, [fetchMonitor])

  // Listen to live SSE check completion events for this monitor
  useEffect(() => {
    if (!monitorId) return
    const stream = new EventSource(`${apiURL}/api/events`)
    stream.onopen = () => setConnected(true)
    stream.onerror = () => setConnected(false)

    const onCheckCompleted = (event: Event) => {
      try {
        const result = JSON.parse((event as MessageEvent<string>).data) as CheckResult
        if (result.monitorId === monitorId) {
          setMonitor((prev) =>
            prev
              ? {
                  ...prev,
                  lastStatus: result.status,
                  lastCheckedAt: result.checkedAt,
                  lastResponseMs: result.responseMs,
                }
              : prev,
          )

          setHistory((prev) => {
            const last = prev[prev.length - 1]
            if (last && last.timestamp === result.checkedAt && last.status === result.status) {
              return prev
            }
            return [
              ...prev,
              {
                ms: result.responseMs,
                status: result.status,
                timestamp: result.checkedAt,
              },
            ].slice(-50)
          })
        }
      } catch {
        // Ignore malformed events
      }
    }

    stream.addEventListener("check.completed", onCheckCompleted)
    return () => {
      stream.removeEventListener("check.completed", onCheckCompleted)
      stream.close()
    }
  }, [monitorId])

  if (loading) {
    return (
      <main className="mx-auto min-h-svh w-full max-w-5xl p-6 md:p-10 flex flex-col items-center justify-center space-y-4">
        <div className="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
        <p className="text-sm text-muted-foreground">Loading monitor insights...</p>
      </main>
    )
  }

  if (error || !monitor) {
    return (
      <main className="mx-auto min-h-svh w-full max-w-5xl p-6 md:p-10 space-y-6">
        <Link href="/" className="text-xs font-semibold text-primary hover:underline">
          ← Back to Dashboard
        </Link>
        <div className="rounded-2xl border border-destructive/30 bg-destructive/10 p-6 text-center">
          <p className="text-sm font-medium text-destructive mb-3">
            {error || "Monitor not found"}
          </p>
          <button
            onClick={() => void fetchMonitor()}
            className="rounded-xl border border-destructive/40 bg-card px-4 py-1.5 text-xs font-semibold text-destructive hover:bg-destructive/10 transition-colors"
          >
            Retry
          </button>
        </div>
      </main>
    )
  }

  const isUp = monitor.lastStatus === "up"
  const isPending = monitor.lastStatus === "pending"
  const statusLabel = isPending ? "Checking" : isUp ? "Operational" : "Down"

  return (
    <main className="mx-auto min-h-svh w-full max-w-5xl p-6 md:p-10 space-y-8">
      {/* Header Bar */}
      <header className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-6">
        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <Link href="/" className="text-xs font-semibold text-primary hover:underline flex items-center gap-1">
              ← Dashboard
            </Link>
            <span className="text-muted-foreground">•</span>
            <span className="text-xs text-muted-foreground font-mono truncate max-w-xs">
              ID: {monitor.id}
            </span>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-3xl font-extrabold tracking-tight">{monitor.name}</h1>
            <span
              className={`rounded-full px-3 py-1 text-xs font-semibold ${
                isPending
                  ? "bg-amber-500/15 text-amber-700 dark:text-amber-300"
                  : isUp
                  ? "bg-emerald-500/15 text-emerald-700 dark:text-emerald-300"
                  : "bg-destructive/15 text-destructive"
              }`}
            >
              {statusLabel}
            </span>
          </div>

          <a
            href={monitor.url}
            target="_blank"
            rel="noreferrer"
            className="text-xs font-mono text-muted-foreground hover:text-primary hover:underline block"
          >
            {monitor.url} ↗
          </a>
        </div>

        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1.5 pr-2 text-xs text-muted-foreground" title="SSE Connection Status">
            <span className={`size-2.5 rounded-full ${connected ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.6)]" : "bg-amber-500 animate-pulse"}`} />
            <span>{connected ? "Live Stream" : "Connecting"}</span>
          </div>

          {user && (
            <>
              <button
                onClick={() => setEditDialogOpen(true)}
                className="rounded-xl border border-border bg-card px-3.5 py-1.5 text-xs font-semibold hover:bg-accent transition-colors"
              >
                Edit
              </button>
              <button
                onClick={() => setDeleteDialogOpen(true)}
                className="rounded-xl border border-destructive/30 bg-destructive/10 px-3.5 py-1.5 text-xs font-semibold text-destructive hover:bg-destructive/20 transition-colors"
              >
                Delete
              </button>
            </>
          )}
        </div>
      </header>

      {/* Latency Smooth Line Graph Section */}
      <section className="space-y-3">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">
          Performance & Latency Analysis
        </h2>
        <LatencyGraph
          history={history}
          currentLatency={monitor.lastResponseMs}
          height={260}
        />
      </section>

      {/* Detailed Metrics Grid */}
      <section className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <div className="rounded-2xl border border-border/80 bg-card p-5 shadow-2xs">
          <span className="text-xs font-medium text-muted-foreground">Last Response Time</span>
          <p className="mt-1 text-2xl font-bold tracking-tight">
            {monitor.lastResponseMs !== undefined ? `${monitor.lastResponseMs} ms` : "—"}
          </p>
        </div>
        <div className="rounded-2xl border border-border/80 bg-card p-5 shadow-2xs">
          <span className="text-xs font-medium text-muted-foreground">Check Frequency</span>
          <p className="mt-1 text-2xl font-bold tracking-tight">
            Every {monitor.intervalSeconds ?? 60}s
          </p>
        </div>
        <div className="rounded-2xl border border-border/80 bg-card p-5 shadow-2xs">
          <span className="text-xs font-medium text-muted-foreground">Last Checked</span>
          <p className="mt-1 text-xs font-semibold text-foreground">
            {formatDate(monitor.lastCheckedAt)}
          </p>
        </div>
        <div className="rounded-2xl border border-border/80 bg-card p-5 shadow-2xs">
          <span className="text-xs font-medium text-muted-foreground">Next Check Scheduled</span>
          <p className="mt-1 text-xs font-semibold text-foreground">
            {formatDate(monitor.nextCheckAt)}
          </p>
        </div>
      </section>

      {/* Dialogs */}
      {user && (
        <>
          <MonitorDialog
            isOpen={editDialogOpen}
            monitorToEdit={monitor}
            onClose={() => setEditDialogOpen(false)}
            onSuccess={() => {
              setEditDialogOpen(false)
              void fetchMonitor()
            }}
          />
          <DeleteDialog
            isOpen={deleteDialogOpen}
            monitor={monitor}
            onClose={() => setDeleteDialogOpen(false)}
            onSuccess={() => {
              setDeleteDialogOpen(false)
              router.push("/")
            }}
          />
        </>
      )}
    </main>
  )
}

function formatDate(value?: string) {
  if (!value) return "Not yet checked"
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? "Not yet checked" : date.toLocaleString()
}
