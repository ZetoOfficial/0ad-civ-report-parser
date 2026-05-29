// Shared chart primitives: Panel renders one Plotly line chart with player-color
// traces, unified MM:SS x-axis ticks, and the hover-x-unified style we want.
import { useEffect, useRef } from "react";
// @ts-expect-error — plotly basic dist ships no types
import * as PlotlyMod from "plotly.js-basic-dist-min";
import type { Analysis } from "../types";
import { colorToCss, formatDuration } from "../utils";
import { useTimeRange, type XRangeSec } from "../timeRange";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const Plotly: any = (PlotlyMod as any).default ?? PlotlyMod;

export function mmss(s: number): string {
  return formatDuration(s * 1000);
}

// Major tick interval in seconds, aiming for 6-10 labels across the range.
function majorTickStep(xMax: number): number {
  const candidates = [30, 60, 120, 180, 300, 600, 900, 1200, 1800, 3600];
  for (const c of candidates) if (xMax / c <= 10) return c;
  return 3600;
}

export interface PlayerSeries {
  player: Analysis["players"][number];
  x: number[];
  y: number[];
}

export interface PhaseMarker {
  t: number;
  label: string;  // short, 1-2 chars (e.g. "T", "C")
  color: string;
  // ySlot ∈ [0,1]: vertical slot from top (0 = topmost). Used to stagger
  // labels per player so they don't collide horizontally when transitions
  // are close in time.
  ySlot: number;
}

// phaseMarkers builds dashed vertical markers for town/city phase transitions
// per player. Label is short ("T"/"C") and colored by player; the full player
// name is identified via the trace legend below the chart.
export function phaseMarkers(analysis: Analysis): PhaseMarker[] {
  const out: PhaseMarker[] = [];
  analysis.players.forEach((p, i) => {
    const pm = analysis.metrics.players[String(p.id)];
    if (!pm) return;
    const c = colorToCss(p.color);
    const slot = i; // one vertical slot per player
    const tTown = pm.phase_timings?.town;
    const tCity = pm.phase_timings?.city;
    if (tTown) out.push({ t: tTown, label: "T", color: c, ySlot: slot });
    if (tCity) out.push({ t: tCity, label: "C", color: c, ySlot: slot });
  });
  return out;
}

export interface PanelProps {
  title: string;
  series: PlayerSeries[];
  xMax: number;
  yLabel?: string;
  height?: number;
  // line.shape passed to Plotly; default linear. Set "hv" for step plots.
  shape?: "linear" | "hv" | "vh";
  // Vertical dashed markers (phase transitions, etc.). Drawn behind the lines.
  markers?: PhaseMarker[];
}

// parseAxisMs parses a Plotly date-axis value (number ms OR ISO/SQL string)
// back to milliseconds. Returns null if unparseable.
function parseAxisMs(v: unknown): number | null {
  if (typeof v === "number") return v;
  if (typeof v === "string") {
    const ms = new Date(v).getTime();
    return Number.isFinite(ms) ? ms : null;
  }
  return null;
}

export function Panel({ title, series, xMax, yLabel, height, shape = "linear", markers }: PanelProps) {
  const ref = useRef<HTMLDivElement | null>(null);
  const { xRange, setXRange } = useTimeRange();
  // Track the range we last *applied* so the relayout listener can detect
  // an echo from our own programmatic update and skip it.
  const lastAppliedRangeRef = useRef<XRangeSec>(null);

  // Effect 1: full plot rebuild when data/styling changes. Listener is
  // attached here, lives for the plot's lifetime.
  useEffect(() => {
    if (!ref.current) return;
    const node = ref.current;
    if (series.length === 0) {
      Plotly.purge(node);
      return;
    }
    const toMs = (sec: number) => sec * 1000;
    const traces = series.map((s) => ({
      name: `${s.player.name} (${s.player.civ})`,
      type: "scatter" as const,
      mode: "lines" as const,
      line: { color: colorToCss(s.player.color), width: 2, shape },
      x: s.x.map(toMs),
      y: s.y,
      hovertemplate: `<b>${s.player.name}</b>: %{y}<extra></extra>`,
    }));
    const shapes: Record<string, unknown>[] = [];
    const annotations: Record<string, unknown>[] = [];
    const slotStep = 0.06;
    for (const m of markers ?? []) {
      shapes.push({
        type: "line",
        xref: "x",
        yref: "paper",
        x0: m.t * 1000,
        x1: m.t * 1000,
        y0: 0,
        y1: 1,
        line: { dash: "dash", width: 1, color: m.color },
      });
      annotations.push({
        x: m.t * 1000,
        y: 1 - m.ySlot * slotStep,
        yref: "paper",
        xanchor: "center",
        yanchor: "top",
        text: m.label,
        showarrow: false,
        font: { size: 10, color: m.color, weight: "bold" },
        bgcolor: "rgba(255,255,255,0.85)",
        borderpad: 1,
      });
    }

    // Resolve effective range. Context overrides per-panel xMax.
    const effMinSec = xRange ? xRange[0] : 0;
    const effMaxSec = xRange ? xRange[1] : xMax;
    lastAppliedRangeRef.current = [effMinSec, effMaxSec];

    const majorMs = majorTickStep(effMaxSec - effMinSec) * 1000;
    Plotly.newPlot(
      node,
      traces,
      {
        margin: { t: 8, r: 8, l: 48, b: 70 },
        xaxis: {
          title: { text: "время" },
          type: "date",
          range: [effMinSec * 1000, effMaxSec * 1000],
          tickformat: "%-M:%S",
          hoverformat: "%-M:%S",
          dtick: majorMs,
          minor: { dtick: 30000, showgrid: true, griddash: "dot" },
        },
        yaxis: { title: { text: yLabel ?? "" }, zeroline: true },
        hovermode: "x unified",
        legend: { orientation: "h", y: -0.35, yanchor: "top" },
        shapes,
        annotations,
        autosize: true,
      },
      {
        responsive: true,
        displayModeBar: "hover",
        modeBarButtonsToRemove: [
          "lasso2d",
          "select2d",
          "toggleSpikelines",
          "hoverClosestCartesian",
          "hoverCompareCartesian",
        ],
      },
    );

    // Sync handler — fires on user zoom/pan (and on our own relayouts).
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const onRelayout = (ev: any) => {
      if (ev["xaxis.autorange"]) {
        setXRange(null);
        return;
      }
      const r0raw = ev["xaxis.range[0]"];
      const r1raw = ev["xaxis.range[1]"];
      if (r0raw === undefined || r1raw === undefined) return;
      const r0ms = parseAxisMs(r0raw);
      const r1ms = parseAxisMs(r1raw);
      if (r0ms === null || r1ms === null) return;
      const newSec: [number, number] = [r0ms / 1000, r1ms / 1000];
      // Echo guard: skip if matches what we just applied.
      const last = lastAppliedRangeRef.current;
      if (last && Math.abs(newSec[0] - last[0]) < 0.5 && Math.abs(newSec[1] - last[1]) < 0.5) return;
      setXRange(newSec);
    };
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (node as any).on("plotly_relayout", onRelayout);

    return () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (node as any).removeAllListeners?.("plotly_relayout");
      Plotly.purge(node);
    };
    // xRange intentionally excluded — handled by Effect 2 via Plotly.relayout
    // to avoid full-rebuild on every range change.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [series, xMax, yLabel, shape, markers, setXRange]);

  // Effect 2: lightweight range sync — fires only when xRange (context)
  // changes. Uses Plotly.relayout so traces aren't re-rendered.
  useEffect(() => {
    if (!ref.current) return;
    const effMinSec = xRange ? xRange[0] : 0;
    const effMaxSec = xRange ? xRange[1] : xMax;
    lastAppliedRangeRef.current = [effMinSec, effMaxSec];
    const majorMs = majorTickStep(effMaxSec - effMinSec) * 1000;
    Plotly.relayout(ref.current, {
      "xaxis.range": [effMinSec * 1000, effMaxSec * 1000],
      "xaxis.dtick": majorMs,
    });
  }, [xRange, xMax]);

  if (series.length === 0) return null;
  const hasMarkers = (markers?.length ?? 0) > 0;
  return (
    <div className="bg-gray-50 rounded-md p-3 border border-gray-100">
      <div className="flex items-baseline justify-between mb-2">
        <h3 className="font-semibold">{title}</h3>
        {hasMarkers && (
          <span className="text-xs text-gray-500">T = town, C = city (фаза, цвет = игрок)</span>
        )}
      </div>
      <div ref={ref} style={{ width: "100%", height: height ?? 220 }} />
    </div>
  );
}
