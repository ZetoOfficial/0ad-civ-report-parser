import type { Analysis } from "../types";
import { Panel, phaseMarkers, type PlayerSeries } from "./charts";

const MILITARY = ["Cavalry", "Infantry", "Champion", "Hero", "Siege", "Ship"];

type Kind = "pop" | "workers" | "army";

function buildSeries(analysis: Analysis, kind: Kind): PlayerSeries[] {
  const out: PlayerSeries[] = [];
  for (const p of analysis.players) {
    const pm = analysis.metrics.players[String(p.id)];
    const s = pm?.sequences;
    if (!s || s.time.length === 0) continue;
    let y: number[] = [];
    if (kind === "pop") {
      y = s.population_count;
    } else if (kind === "workers") {
      const trained = s.units_trained.Worker ?? [];
      const lost = s.units_lost.Worker ?? [];
      y = trained.map((v, i) => Math.max(0, v - (lost[i] ?? 0)));
    } else {
      // army = sum_classes(trained - lost)
      y = s.time.map((_, i) => {
        let total = 0;
        for (const cls of MILITARY) {
          total += (s.units_trained[cls]?.[i] ?? 0) - (s.units_lost[cls]?.[i] ?? 0);
        }
        return Math.max(0, total);
      });
    }
    if (y.length === 0) continue;
    out.push({ player: p, x: s.time, y });
  }
  return out;
}

export function PopulationChart({ analysis }: { analysis: Analysis }) {
  const pop = buildSeries(analysis, "pop");
  const workers = buildSeries(analysis, "workers");
  const army = buildSeries(analysis, "army");

  if (pop.length === 0) {
    return <p className="text-gray-500 italic">no sequences data</p>;
  }

  const xMax = Math.max(
    Math.ceil(analysis.game.duration_ms / 1000),
    ...pop.flatMap((s) => s.x),
  );
  const markers = phaseMarkers(analysis);

  return (
    <div className="space-y-3">
      <Panel title="Популяция (всего)" series={pop} xMax={xMax} yLabel="юнитов" markers={markers} />
      <Panel title="Работяги (живых)" series={workers} xMax={xMax} yLabel="юнитов" markers={markers} />
      <Panel title="Армия (живых, без работяг)" series={army} xMax={xMax} yLabel="юнитов" markers={markers} />
    </div>
  );
}
