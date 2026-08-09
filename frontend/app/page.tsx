"use client"

import { useState } from "react"
import Link from "next/link"
import { useAuth } from "@/context/auth-context"
import { MonitorList } from "@/components/monitor-list"
import { MonitorDialog } from "@/components/monitor-dialog"
import { AuthModal } from "@/components/auth-modal"

export default function DashboardPage() {
  const { user, isLoading } = useAuth()
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [authModalOpen, setAuthModalOpen] = useState(false)
  const [authMode, setAuthMode] = useState<"login" | "register">("register")
  const [refreshKey, setRefreshKey] = useState(0)

  const openAuth = (mode: "login" | "register") => {
    setAuthMode(mode)
    setAuthModalOpen(true)
  }

  const triggerRefresh = () => {
    setRefreshKey((prev) => prev + 1)
  }

  if (isLoading) {
    return (
      <main className="mx-auto min-h-[70vh] w-full max-w-6xl p-6 md:p-10 flex items-center justify-center">
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          <svg className="size-5 animate-spin text-primary" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          Loading your session...
        </div>
      </main>
    )
  }

  return (
    <main className="mx-auto min-h-svh w-full max-w-6xl p-6 md:p-10 space-y-8">
      {user ? (
        <>
          <header className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-6">
            <div>
              <p className="text-xs font-semibold text-primary uppercase tracking-wider">Dashboard</p>
              <h1 className="text-3xl font-bold tracking-tight mt-1">
                Welcome back, {user.firstName}!
              </h1>
              <p className="mt-1 text-sm text-muted-foreground">
                Real-time latency metrics and automated uptime monitoring.
              </p>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setCreateDialogOpen(true)}
                className="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-md hover:bg-primary/90 transition-colors"
              >
                <svg className="size-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                </svg>
                Add Monitor
              </button>
            </div>
          </header>

          <MonitorList refreshTrigger={refreshKey} />

          <MonitorDialog
            isOpen={createDialogOpen}
            onClose={() => setCreateDialogOpen(false)}
            onSuccess={() => {
              setCreateDialogOpen(false)
              triggerRefresh()
            }}
          />
        </>
      ) : (
        <div className="py-12 md:py-20 text-center max-w-3xl mx-auto space-y-8">
          <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-3.5 py-1 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
            <span className="size-2 rounded-full bg-emerald-500 animate-pulse" />
            Go-Powered High Frequency Uptime Check Engine
          </div>

          <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight">
            Instant Uptime Monitoring & Real-time Alerts
          </h1>

          <p className="text-base sm:text-lg text-muted-foreground leading-relaxed">
            Monitor your HTTP APIs, web applications, and endpoints with sub-minute checks. Receive instant Server-Sent Events live stream updates and status visibility.
          </p>

          <div className="flex flex-wrap items-center justify-center gap-4 pt-4">
            <button
              onClick={() => openAuth("register")}
              className="rounded-xl bg-primary px-6 py-3 text-sm font-bold text-primary-foreground shadow-lg hover:bg-primary/90 transition-all hover:scale-102"
            >
              Start Monitoring Free
            </button>
            <Link
              href="/status"
              className="rounded-xl border border-border bg-card px-6 py-3 text-sm font-semibold hover:bg-accent transition-colors"
            >
              View Public Status Page →
            </Link>
          </div>

          <div className="pt-12">
            <div className="text-left border border-border/60 rounded-2xl p-6 bg-card shadow-lg">
              <div className="flex items-center justify-between mb-4 border-b border-border pb-3">
                <h3 className="font-semibold text-sm">Public Status Preview</h3>
                <span className="text-xs text-muted-foreground">Live Public Endpoints</span>
              </div>
              <MonitorList publicView />
            </div>
          </div>
        </div>
      )}

      <AuthModal
        isOpen={authModalOpen}
        initialMode={authMode}
        onClose={() => setAuthModalOpen(false)}
      />
    </main>
  )
}
