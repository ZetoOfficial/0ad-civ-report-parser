import { useEffect, useRef } from "react";
// @ts-expect-error — plotly.js-dist-min ships no types
import * as PlotlyMod from "plotly.js-dist-min";
import type { Analysis } from "../types";

// Vite wraps the CJS plotly bundle inconsistently — sometimes the namespace
// itself is the Plotly object, sometimes it's `{default: Plotly}`. Unwrap.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const Plotly: any = (PlotlyMod as any).default ?? PlotlyMod;

interface Props { analysis: Analysis }

const CATEGORIES = ["military", "build", "research", "economy", "other"] as const;

export function DensityChart({ analysis }: Props) {
  const ref = useRef<HTMLDivElement | null>(null);
  const bins = analysis.metrics.action_density;
  const hasBins = bins && bins.length > 0;

  useEffect(() => {
    if (!ref.current || !hasBins) return;
    const node = ref.current;

    const x = bins.map((b) => b.t_sec);
    const traces = CATEGORIES.map((cat) => ({
      name: cat,
      type: "bar" as const,
      x,
      y: bins.map((b) => b.counts[cat] ?? 0),
    }));

    const shapes: Record<string, unknown>[] = [];
    const annotations: Record<string, unknown>[] = [];
    for (const m of Object.values(analysis.metrics.players)) {
      for (const [name, t] of Object.entries(m.phase_timings)) {
        shapes.push({
          type: "line", xref: "x", yref: "paper",
          x0: t, x1: t, y0: 0, y1: 1,
          line: { dash: "dash", width: 1, color: "#555" },
        });
        annotations.push({
          x: t, y: 1, yref: "paper",
          text: name, showarrow: false,
          font: { size: 10, color: "#555" },
        });
      }
      for (const e of m.engagements) {
        if (e.peak_units < 5) continue;
        shapes.push({
          type: "line", xref: "x", yref: "paper",
          x0: e.t_start_sec, x1: e.t_start_sec, y0: 0, y1: 1,
          line: { dash: "solid", width: Math.min(1 + Math.log2(e.peak_units), 4), color: "rgba(220,0,0,0.5)" },
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
