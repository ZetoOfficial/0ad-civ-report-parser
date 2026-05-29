// Per-resource time-series: current balance + cumulative gathered, per resource.
// "Current balance" is the answer to "был ли я food-locked в 14 мин".
import { useState } from "react";
import type { Analysis, ResourceSeries } from "../types";
import { Panel, phaseMarkers, type PlayerSeries } from "./charts";

const RESOURCES = ["food", "wood", "stone", "metal"] as const;
type Resource = (typeof RESOURCES)[number];

const RU_LABEL: Record<Resource, string> = {
  food: "Еда",
  wood: "Дерево",
  stone: "Камень",
  metal: "Металл",
};

// Approximate per-worker base gather rate (resource / sec) in 0 A.D. R28
// without tech upgrades. Used to back out worker count from gather-rate.
// Source: templates citizen-female / citizen-soldier baseline ResourceGatherer.
const PER_WORKER_RATE: Record<Resource, number> = {
  food: 0.5,   // mixed farming/berries/hunt; farming is the steady-state mode
  wood: 0.7,   // chopping trees
  stone: 0.45, // mining
  metal: 0.45, // mining
};

// Sliding window for rate derivative. 60s ≈ 2 samples at 30s spacing → smooths
// out the spiky single-sample noise without lagging too hard.
const RATE_WINDOW_SEC = 60;

type Metric = "count" | "gathered" | "used" | "net_market" | "workers";

const METRIC_LABEL: Record<Metric, string> = {
  count: "Текущий запас",
  gathered: "Добыто (накопительно)",
  used: "Потрачено (накопительно)",
  net_market: "Чистая торговля (продал − купил)",
  workers: "≈ Работяги (оценка)",
};

function computeRate(time: number[], cum: number[]): number[] {
  return time.map((t, i) => {
    let j = i;
    while (j > 0 && t - time[j - 1] < RATE_WINDOW_SEC) j--;
    const dt = t - time[j];
    if (dt <= 0) return 0;
    return Math.max(0, (cum[i] - cum[j]) / dt);
  });
}

function pickSource(s: Analysis["metrics"]["players"][string]["sequences"], metric: Metric): ResourceSeries {
  if (!s) return {} as ResourceSeries;
  if (metric === "count") return s.resources_count;
  if (metric === "gathered" || metric === "workers") return s.resources_gathered;
  if (metric === "used") return s.resources_used;
  return {} as ResourceSeries;
}

function buildSeries(analysis: Analysis, resource: Resource, metric: Metric): PlayerSeries[] {
  const out: PlayerSeries[] = [];
  for (const p of analysis.players) {
    const pm = analysis.metrics.players[String(p.id)];
    const s = pm?.sequences;
    if (!s || s.time.length === 0) continue;
    let y: number[] = [];
    if (metric === "net_market") {
      const sold = s.resources_sold?.[resource] ?? [];
      const bought = s.resources_bought?.[resource] ?? [];
      y = s.time.map((_, i) => (sold[i] ?? 0) - (bought[i] ?? 0));
    } else if (metric === "workers") {
      const cum = s.resources_gathered?.[resource] ?? [];
      if (cum.length === 0) continue;
      const rate = computeRate(s.time, cum);
      y = rate.map((r) => r / PER_WORKER_RATE[resource]);
    } else {
      const src = pickSource(s, metric);
      y = src?.[resource] ?? [];
    }
    if (!y || y.length === 0) continue;
    out.push({ player: p, x: s.time, y });
  }
  return out;
}

export function ResourceChart({ analysis }: { analysis: Analysis }) {
  const [metric, setMetric] = useState<Metric>("count");

  const xMax = (() => {
    let m = Math.ceil(analysis.game.duration_ms / 1000);
    for (const pm of Object.values(analysis.metrics.players)) {
      const t = pm?.sequences?.time;
      if (t && t.length > 0) m = Math.max(m, t[t.length - 1]);
    }
    return m;
  })();
  const markers = phaseMarkers(analysis);

  // Skip resources nobody used to keep the page tight.
  const visible = RESOURCES.filter((r) => buildSeries(analysis, r, metric).length > 0);

  if (visible.length === 0) {
    return <p className="text-gray-500 italic">no resource data</p>;
  }

  const yLabel =
    metric === "count" ? "запас" :
    metric === "workers" ? "≈ юнитов" :
    "ресурсов";

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap gap-2 text-sm">
        <span className="text-gray-600 self-center">метрика:</span>
        {(Object.keys(METRIC_LABEL) as Metric[]).map((m) => (
          <button
            key={m}
            type="button"
            onClick={() => setMetric(m)}
            className={
              "px-2 py-1 rounded border " +
              (m === metric
                ? "bg-blue-600 text-white border-blue-600"
                : "bg-white text-gray-700 border-gray-300 hover:border-gray-400")
            }
          >
            {METRIC_LABEL[m]}
          </button>
        ))}
      </div>
      {metric === "workers" && (
        <p className="text-xs text-gray-500">
          Оценка из скорости добычи ({RATE_WINDOW_SEC}s окно) ÷ базовой rate
          одного работяги (food {PER_WORKER_RATE.food}, wood {PER_WORKER_RATE.wood},
          stone {PER_WORKER_RATE.stone}, metal {PER_WORKER_RATE.metal} /sec).
          Точность ±20%: техи и не-базовые юниты (hunt, fishing, women на ферме)
          сдвигают rate. Поле «сколько юнитов на ресурсе» в replay-данных
          отсутствует — это прокси.
        </p>
      )}
      {visible.map((r) => (
        <Panel
          key={r}
          title={`${RU_LABEL[r]} — ${METRIC_LABEL[metric].toLowerCase()}`}
          series={buildSeries(analysis, r, metric)}
          xMax={xMax}
          yLabel={yLabel}
          markers={markers}
        />
      ))}
    </div>
  );
}
