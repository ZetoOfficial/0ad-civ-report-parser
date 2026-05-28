import { useEffect, useRef } from "react";
// @ts-expect-error — plotly basic dist ships no types
import * as PlotlyMod from "plotly.js-basic-dist-min";
import type { Analysis, Player } from "../types";
import { colorToCss } from "../utils";

// Vite wraps the CJS plotly bundle inconsistently — sometimes the namespace
// itself is the Plotly object, sometimes it's `{default: Plotly}`. Unwrap.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const Plotly: any = (PlotlyMod as any).default ?? PlotlyMod;

interface Props { analysis: Analysis }

const CATEGORIES = ["military", "build", "research", "economy", "other"] as const;

// Players whose phase_timings agree within this many seconds collapse into
// one annotation labelled with everyone's name.
const PHASE_COLLAPSE_SEC = 15;

export function DensityChart({ analysis }: Props) {
  const ref = useRef<HTMLDivElement | null>(null);
  const bins = analysis.metrics.action_density;
  const hasBins = bins && bins.length > 0;

  useEffect(() => {
    if (!ref.current || !hasBins) return;
    const node = ref.current;

    const playerByID: Record<number, Player> = {};
    for (const p of analysis.players) playerByID[p.id] = p;

    const x = bins.map((b) => b.t_sec);
    const traces = CATEGORIES.map((cat) => ({
      name: cat,
      type: "bar" as const,
      x,
      y: bins.map((b) => b.counts[cat] ?? 0),
    }));

    // Phase markers: dedupe by (phase, ~time), label with player names.
    type PhaseHit = { phase: string; t: number; players: Player[] };
    const phaseHits: PhaseHit[] = [];
    for (const [pidStr, m] of Object.entries(analysis.metrics.players)) {
      const player = playerByID[Number(pidStr)];
      if (!player) continue;
      for (const [phase, t] of Object.entries(m.phase_timings ?? {})) {
        const existing = phaseHits.find(
          (h) => h.phase === phase && Math.abs(h.t - t) <= PHASE_COLLAPSE_SEC,
        );
        if (existing) existing.players.push(player);
        else phaseHits.push({ phase, t, players: [player] });
      }
    }

    const shapes: Record<string, unknown>[] = [];
    const annotations: Record<string, unknown>[] = [];

    for (const h of phaseHits) {
      // Single dashed line; color from the first player; label lists all.
      const c = colorToCss(h.players[0].color);
      shapes.push({
        type: "line", xref: "x", yref: "paper",
        x0: h.t, x1: h.t, y0: 0, y1: 1,
        line: { dash: "dash", width: 1, color: c },
      });
      annotations.push({
        x: h.t, y: 1, yref: "paper", xanchor: "left",
        text: `${h.phase.replace(/^phase_/, "")} (${h.players.map((p) => p.name).join(",")})`,
        showarrow: false,
        font: { size: 10, color: c },
      });
    }

    // Engagements: line colored by initiating player.
    for (const [pidStr, m] of Object.entries(analysis.metrics.players)) {
      const player = playerByID[Number(pidStr)];
      if (!player) continue;
      const c = colorToCss(player.color);
      for (const e of m.engagements ?? []) {
        if (e.peak_units < 5) continue;
        shapes.push({
          type: "line", xref: "x", yref: "paper",
          x0: e.t_start_sec, x1: e.t_start_sec, y0: 0, y1: 1,
          line: { dash: "solid", width: Math.min(1 + Math.log2(e.peak_units), 4), color: c, opacity: 0.5 },
        });
      }
    }

    Plotly.newPlot(node, traces, {
      barmode: "stack",
      margin: { t: 24, r: 16, l: 40, b: 32 },
      xaxis: { title: { text: "сек" } },
      yaxis: { title: { text: "команд / 30 сек" } },
      shapes,
      annotations,
      legend: { orientation: "h", y: -0.2 },
      autosize: true,
    }, { responsive: true, displayModeBar: false });

    return () => { Plotly.purge(node); };
  }, [analysis, bins, hasBins]);

  if (!hasBins) {
    return <p className="text-gray-500 italic">no command bins</p>;
  }
  return <div ref={ref} style={{ width: "100%", height: 360 }} />;
}
