"use client"

import React, { useState } from "react"
import { deleteMonitor, type Monitor } from "@/lib/monitoring"

type DeleteDialogProps = {
  isOpen: boolean
  monitor: Monitor | null
  onClose: () => void
  onSuccess: () => void
}

export function DeleteDialog({
  isOpen,
  monitor,
  onClose,
  onSuccess,
}: DeleteDialogProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isOpen || !monitor) return null

  const handleDelete = async () => {
    setError(null)
    setLoading(true)
    try {
      await deleteMonitor(monitor.id)
      onSuccess()
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete monitor.")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-in fade-in duration-200">
      <div
        className="w-full max-w-sm rounded-2xl border border-border bg-card p-6 shadow-2xl animate-in zoom-in-95 duration-200"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-border pb-4 mb-4">
          <h2 className="text-lg font-semibold tracking-tight text-destructive">
            Delete Monitor
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

        <p className="text-sm text-muted-foreground mb-6">
          Are you sure you want to delete <strong className="text-foreground">{monitor.name}</strong>? This action cannot be undone and monitoring history will be lost.
        </p>

        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl border border-border px-4 py-2 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={loading}
            onClick={handleDelete}
            className="rounded-xl bg-destructive px-4 py-2 text-sm font-semibold text-destructive-foreground shadow-md hover:bg-destructive/90 disabled:opacity-50 transition-all"
          >
            {loading ? "Deleting..." : "Delete Monitor"}
          </button>
        </div>
      </div>
    </div>
  )
}
