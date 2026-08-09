"use client"

import React, { useEffect, useState } from "react"
import {
  createMonitor,
  updateMonitor,
  type Monitor,
} from "@/lib/monitoring"

type MonitorDialogProps = {
  isOpen: boolean
  monitorToEdit?: Monitor | null
  onClose: () => void
  onSuccess: () => void
}

const INTERVAL_PRESETS = [
  { label: "10s", value: 10 },
  { label: "30s", value: 30 },
  { label: "1 min", value: 60 },
  { label: "5 min", value: 300 },
  { label: "10 min", value: 600 },
]

export function MonitorDialog({
  isOpen,
  monitorToEdit,
  onClose,
  onSuccess,
}: MonitorDialogProps) {
  const isEditing = Boolean(monitorToEdit)
  const [name, setName] = useState("")
  const [url, setUrl] = useState("")
  const [intervalSeconds, setIntervalSeconds] = useState(60)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (monitorToEdit) {
      setName(monitorToEdit.name)
      setUrl(monitorToEdit.url)
      setIntervalSeconds(monitorToEdit.intervalSeconds ?? 60)
    } else {
      setName("")
      setUrl("")
      setIntervalSeconds(60)
    }
    setError(null)
  }, [monitorToEdit, isOpen])

  if (!isOpen) return null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)

    try {
      if (!url.startsWith("http://") && !url.startsWith("https://")) {
        throw new Error("URL must begin with http:// or https://")
      }

      if (isEditing && monitorToEdit) {
        await updateMonitor(monitorToEdit.id, {
          name,
          url,
          intervalSeconds,
        })
      } else {
        await createMonitor({
          name,
          url,
          intervalSeconds,
        })
      }
      onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save monitor.")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-200">
      <div
        className="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-border pb-4 mb-5">
          <h2 className="text-lg font-semibold tracking-tight">
            {isEditing ? "Edit Monitor" : "Add New Monitor"}
          </h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
          >
            <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded-xl border border-destructive/30 bg-destructive/10 p-3 text-xs font-medium text-destructive">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">
              Monitor Name
            </label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="API Gateway / Landing Page"
              className="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">
              Endpoint URL
            </label>
            <input
              type="url"
              required
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://api.example.com/health"
              className="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">
              Check Interval (seconds)
            </label>
            <div className="flex items-center gap-2 mb-2">
              {INTERVAL_PRESETS.map((preset) => (
                <button
                  key={preset.value}
                  type="button"
                  onClick={() => setIntervalSeconds(preset.value)}
                  className={`rounded-lg px-2.5 py-1 text-xs font-medium transition-all ${
                    intervalSeconds === preset.value
                      ? "bg-primary text-primary-foreground shadow-xs"
                      : "bg-muted text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {preset.label}
                </button>
              ))}
            </div>
            <input
              type="number"
              min={10}
              max={86400}
              required
              value={intervalSeconds}
              onChange={(e) => setIntervalSeconds(Number(e.target.value))}
              className="w-full rounded-xl border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-xl border border-border px-4 py-2 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-md hover:bg-primary/90 disabled:opacity-50 transition-all"
            >
              {loading ? "Saving..." : isEditing ? "Save Changes" : "Create Monitor"}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
