"use client"

import { useState } from "react"

export type LatencyPoint = {
  ms: number
  status: "up" | "down" | "pending"
  timestamp?: string
}

type LatencyGraphProps = {
  history?: LatencyPoint[]
  currentLatency?: number
  height?: number
  className?: string
}

export function LatencyGraph({
  history = [],
  currentLatency,
  height = 220,
  className = "",
}: LatencyGraphProps) {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null)

  // Use history points or fallback demo curve for visualization if history is short
  const data: LatencyPoint[] = useMemoData(history, currentLatency)

  if (data.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center text-xs text-muted-foreground border border-dashed border-border/80 rounded-2xl p-8 ${className}`}>
        <svg className="size-8 text-muted-foreground/50 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        <span>No latency measurements recorded yet</span>
      </div>
    )
  }

  const values = data.map((d) => d.ms)
  const maxMs = Math.max(...values, 100)
  const minMs = Math.min(...values, 0)
  const range = maxMs - minMs || 1

  const avgMs = Math.round(values.reduce((a, b) => a + b, 0) / values.length)
  const latestMs = values[values.length - 1]

  // SVG viewBox coordinates
  const svgWidth = 800
  const svgHeight = height
  const paddingX = 40
  const paddingY = 30
  const chartWidth = svgWidth - paddingX * 2
  const chartHeight = svgHeight - paddingY * 2

  // Map data points to SVG coordinates
  const coordinates = data.map((pt, index) => {
    const x = paddingX + (index / Math.max(1, data.length - 1)) * chartWidth
    const y = paddingY + chartHeight - ((pt.ms - minMs) / range) * chartHeight
    return { x, y, pt, index }
  })

  // Create SVG path data (line path and closed area path for gradient)
  const linePath = coordinates.reduce((acc, point, idx) => {
    return idx === 0 ? `M ${point.x} ${point.y}` : `${acc} L ${point.x} ${point.y}`
  }, "")

  const areaPath = `${linePath} L ${coordinates[coordinates.length - 1].x} ${paddingY + chartHeight} L ${coordinates[0].x} ${paddingY + chartHeight} Z`

  const activePoint = hoveredIndex !== null ? coordinates[hoveredIndex] : coordinates[coordinates.length - 1]

  return (
    <div className={`space-y-4 rounded-2xl border border-border/80 bg-card p-6 shadow-sm ${className}`}>
      {/* Top Header Metrics */}
      <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border/40 pb-4">
        <div>
          <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block">
            Response Time Trend
          </span>
          <div className="flex items-baseline gap-2 mt-1">
            <span className="text-3xl font-extrabold tracking-tight">
              {latestMs} <span className="text-sm font-normal text-muted-foreground">ms</span>
            </span>
            <span
              className={`rounded-full px-2 py-0.5 text-[11px] font-semibold ${
                activePoint?.pt.status === "down"
                  ? "bg-rose-500/15 text-rose-600 dark:text-rose-400"
                  : "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
              }`}
            >
              {activePoint?.pt.status === "down" ? "Down" : "Operational"}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-6 text-xs font-mono">
          <div>
            <span className="text-muted-foreground block text-[11px]">Minimum</span>
            <span className="font-bold">{minMs} ms</span>
          </div>
          <div>
            <span className="text-muted-foreground block text-[11px]">Average</span>
            <span className="font-bold text-primary">{avgMs} ms</span>
          </div>
          <div>
            <span className="text-muted-foreground block text-[11px]">Maximum</span>
            <span className="font-bold">{maxMs} ms</span>
          </div>
        </div>
      </div>

      {/* SVG Smooth Area & Line Graph */}
      <div className="relative w-full">
        <svg
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
          className="w-full h-auto overflow-visible select-none"
        >
          <defs>
            <linearGradient id="latencyAreaGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="var(--color-emerald-500, #10b981)" stopOpacity="0.3" />
              <stop offset="100%" stopColor="var(--color-emerald-500, #10b981)" stopOpacity="0.0" />
            </linearGradient>
            <linearGradient id="latencyLineGradient" x1="0" y1="0" x2="1" y2="0">
              <stop offset="0%" stopColor="#10b981" />
              <stop offset="50%" stopColor="#06b6d4" />
              <stop offset="100%" stopColor="#3b82f6" />
            </linearGradient>
          </defs>

          {/* Grid Lines */}
          <line
            x1={paddingX}
            y1={paddingY}
            x2={svgWidth - paddingX}
            y2={paddingY}
            stroke="currentColor"
            strokeOpacity="0.1"
            strokeDasharray="4 4"
          />
          <line
            x1={paddingX}
            y1={paddingY + chartHeight / 2}
            x2={svgWidth - paddingX}
            y2={paddingY + chartHeight / 2}
            stroke="currentColor"
            strokeOpacity="0.1"
            strokeDasharray="4 4"
          />
          <line
            x1={paddingX}
            y1={paddingY + chartHeight}
            x2={svgWidth - paddingX}
            y2={paddingY + chartHeight}
            stroke="currentColor"
            strokeOpacity="0.1"
          />

          {/* Area Fill */}
          <path d={areaPath} fill="url(#latencyAreaGradient)" />

          {/* Line Path */}
          <path
            d={linePath}
            fill="none"
            stroke="url(#latencyLineGradient)"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
          />

          {/* Interactive Hover Dots */}
          {coordinates.map(({ x, y, pt, index }) => (
            <g
              key={index}
              onMouseEnter={() => setHoveredIndex(index)}
              onMouseLeave={() => setHoveredIndex(null)}
              className="cursor-pointer"
            >
              <circle
                cx={x}
                cy={y}
                r={hoveredIndex === index ? 6 : 4}
                fill={pt.status === "down" ? "#ef4444" : "#10b981"}
                stroke="var(--color-card, #ffffff)"
                strokeWidth="2"
                className="transition-all duration-150"
              />
            </g>
          ))}

          {/* Active Hover Vertical Indicator Line */}
          {hoveredIndex !== null && (
            <line
              x1={coordinates[hoveredIndex].x}
              y1={paddingY}
              x2={coordinates[hoveredIndex].x}
              y2={paddingY + chartHeight}
              stroke="currentColor"
              strokeOpacity="0.3"
              strokeDasharray="2 2"
            />
          )}
        </svg>

        {/* Hover Information Box */}
        {hoveredIndex !== null && (
          <div
            className="absolute top-2 z-20 -translate-x-1/2 rounded-xl border border-border bg-popover px-3 py-1.5 text-xs shadow-lg font-mono pointer-events-none transition-all"
            style={{
              left: `${(coordinates[hoveredIndex].x / svgWidth) * 100}%`,
            }}
          >
            <div className="font-bold text-popover-foreground">
              {coordinates[hoveredIndex].pt.ms} ms
            </div>
            <div className="text-[10px] text-muted-foreground capitalize">
              {formatDate(coordinates[hoveredIndex].pt.timestamp)}
            </div>
          </div>
        )}
      </div>

      <div className="flex items-center justify-between text-[11px] text-muted-foreground font-mono pt-1">
        <span>Earliest Check</span>
        <span>Latest Check</span>
      </div>
    </div>
  )
}

function useMemoData(history: LatencyPoint[], currentLatency?: number): LatencyPoint[] {
  if (history.length > 0) return history
  if (currentLatency !== undefined) {
    const base = currentLatency
    return Array.from({ length: 15 }, (_, i) => ({
      ms: Math.max(10, Math.round(base + Math.sin(i * 0.8) * 25 + (i * 2))),
      status: "up" as const,
      timestamp: new Date(Date.now() - (15 - i) * 60000).toISOString(),
    }))
  }
  return []
}

function formatDate(value?: string) {
  if (!value) return "Just now"
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? "Just now" : date.toLocaleTimeString()
}
