import createPlotlyComponent from "react-plotly.js/factory";
// @ts-expect-error — plotly.js-dist-min ships no types; using @types/plotly.js indirectly
import Plotly from "plotly.js-dist-min";
import type { Analysis } from "../types";

// Factory pattern bypasses Vite's ESM interop issue where the default export
// of "react-plotly.js" comes through as `{default: Component}` instead of the
// Component itself.
const Plot = createPlotlyComponent(Plotly);

interface Props { analysis: Analysis }

const CATEGORIES = ["military", "build", "research", "economy", "other"] as const;

export function DensityChart({ analysis }: Props) {
  const bins = analysis.metrics.action_density;
  if (!bins || bins.length === 0) {
    return <p className="text-gray-500 italic">no command bins</p>;
  }

  const x = bins.map((b) => b.t_sec);
  const traces = CATEGORIES.map((cat) => ({
    name: cat,
    type: "bar" as const,
    x,
    y: bins.map((b) => b.counts[cat] ?? 0),
  }));

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const phaseShapes: any[] = [];
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const phaseAnnotations: any[] = [];
  for (const m of Object.values(analysis.metrics.players)) {
    for (const [name, t] of Object.entries(m.phase_timings)) {
      phaseShapes.push({
        type: "line", xref: "x", yref: "paper",
        x0: t, x1: t, y0: 0, y1: 1,
        line: { dash: "dash", width: 1, color: "#555" },
      });
      phaseAnnotations.push({
        x: t, y: 1, yref: "paper",
        text: name, showarrow: false,
        font: { size: 10, color: "#555" },
      });
    }
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const engShapes: any[] = [];
  for (const m of Object.values(analysis.metrics.players)) {
    for (const e of m.engagements) {
      if (e.peak_units < 5) continue;
      engShapes.push({
        type: "line", xref: "x", yref: "paper",
        x0: e.t_start_sec, x1: e.t_start_sec, y0: 0, y1: 1,
        line: { dash: "solid", width: Math.min(1 + Math.log2(e.peak_units), 4), color: "rgba(220,0,0,0.5)" },
      });
    }
  }

  return (
    <Plot
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data={traces as any}
      layout={{
        barmode: "stack",
        margin: { t: 24, r: 16, l: 40, b: 32 },
        xaxis: { title: { text: "сек" } },
        yaxis: { title: { text: "команд / 30 сек" } },
        shapes: [...phaseShapes, ...engShapes],
        annotations: phaseAnnotations,
        legend: { orientation: "h", y: -0.2 },
        autosize: true,
      }}
      style={{ width: "100%", height: 360 }}
      useResizeHandler
      config={{ displayModeBar: false }}
    />
  );
}
