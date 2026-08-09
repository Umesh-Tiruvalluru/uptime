import Link from "next/link"
import { MonitorList } from "@/components/monitor-list"

export const metadata = {
  title: "System Status | Uptime",
  description: "Live uptime and performance status for all public services.",
}

export default function StatusPage() {
  return (
    <main className="mx-auto min-h-svh w-full max-w-6xl p-6 md:p-10 space-y-8">
      <header className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/40 pb-6">
        <div>
          <div className="flex items-center gap-2">
            <span className="relative flex size-3">
              <span className="absolute inline-flex size-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex size-3 rounded-full bg-emerald-500"></span>
            </span>
            <p className="text-xs font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider">
              All Systems Operational
            </p>
          </div>
          <h1 className="text-3xl font-bold tracking-tight mt-1">System Status</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Live availability and response times for public services.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Link
            className="rounded-xl border border-border bg-card px-4 py-2 text-sm font-semibold hover:bg-accent transition-colors"
            href="/"
          >
            ← Back to Dashboard
          </Link>
        </div>
      </header>

      <MonitorList publicView />
    </main>
  )
}
