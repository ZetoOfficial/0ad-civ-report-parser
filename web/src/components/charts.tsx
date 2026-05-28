// Shared chart primitives: Panel renders one Plotly line chart with player-color
// traces, unified MM:SS x-axis ticks, and the hover-x-unified style we want.
import { useEffect, useRef } from "react";
// @ts-expect-error — plotly basic dist ships no types
import * as PlotlyMod from "plotly.js-basic-dist-min";
import type { Player } from "../types";
import { colorToCss, formatDuration } from "../utils";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const Plotly: any = (PlotlyMod as any).default ?? PlotlyMod;

export function mmss(s: number): string {
  return formatDuration(s * 1000);
}

function tickStep(xMax: number): number {
  const candidates = [30, 60, 120, 180, 300, 600, 900, 1200];
  for (const c of candidates) if (xMax / c <= 10) return c;
  return 1800;
}

export function makeTicks(xMax: number): { tickvals: number[]; ticktext: string[] } {
  const step = tickStep(xMax);
  const tickvals: number[] = [];
  for (let t = 0; t <= xMax; t += step) tickvals.push(t);
  return { tickvals, ticktext: tickvals.map(mmss) };
}

export interface PlayerSeries {
  player: Player;
  x: number[];
  y: number[];
}

export interface PanelProps {
  title: string;
  series: PlayerSeries[];
  xMax: number;
  yLabel?: string;
  height?: number;
  // line.shape passed to Plotly; default linear. Set "hv" for step plots.
  shape?: "linear" | "hv" | "vh";
}

export function Panel({ title, series, xMax, yLabel, height, shape = "linear" }: PanelProps) {
  const ref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!ref.current) return;
    const node = ref.current;
    if (series.length === 0) {
      Plotly.purge(node);
      return;
    }
    const traces = series.map((s) => ({
      name: `${s.player.name} (${s.player.civ})`,
      type: "scatter" as const,
      mode: "lines" as const,
      line: { color: colorToCss(s.player.color), width: 2, shape },
      x: s.x,
      y: s.y,
      customdata: s.x.map(mmss),
      hovertemplate: `<b>${s.player.name}</b>: %{y} @ %{customdata}<extra></extra>`,
    }));
    const { tickvals, ticktext } = makeTicks(xMax);
    Plotly.newPlot(
      node,
      traces,
      {
        margin: { t: 8, r: 8, l: 48, b: 28 },
        xaxis: {
          title: { text: "время" },
          range: [0, xMax],
          tickmode: "array",
          tickvals,
          ticktext,
        },
        yaxis: { title: { text: yLabel ?? "" }, zeroline: true },
        hovermode: "x unified",
        legend: { orientation: "h", y: -0.25 },
        autosize: true,
      },
      { responsive: true, displayModeBar: false },
    );
    return () => {
      Plotly.purge(node);
    };
  }, [series, xMax, yLabel, shape]);

  if (series.length === 0) return null;
  return (
    <div className="bg-gray-50 rounded-md p-3 border border-gray-100">
      <h3 className="font-semibold mb-2">{title}</h3>
      <div ref={ref} style={{ width: "100%", height: height ?? 220 }} />
    </div>
  );
}
