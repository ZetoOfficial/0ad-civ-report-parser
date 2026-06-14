import type { Analysis } from "../types";
import { Panel, phaseMarkers, type PlayerSeries } from "./charts";

type Kind = "killed" | "lost" | "net_value";

function buildSeries(analysis: Analysis, kind: Kind): PlayerSeries[] {
  const out: PlayerSeries[] = [];
  for (const p of analysis.players) {
    const pm = analysis.metrics.players[String(p.id)];
    const s = pm?.sequences;
    if (!s || s.time.length === 0) continue;
    let y: number[] = [];
    if (kind === "killed") {
      y = s.enemy_units_killed.Unit ?? [];
    } else if (kind === "lost") {
      y = s.units_lost.Unit ?? [];
    } else {
      // net_value = enemy_units_killed_value - units_lost_value
      const k = s.enemy_units_killed_value ?? [];
      const l = s.units_lost_value ?? [];
      y = s.time.map((_, i) => Math.round((k[i] ?? 0) - (l[i] ?? 0)));
    }
    if (y.length === 0) continue;
    out.push({ player: p, x: s.time, y });
  }
  return out;
}

export function CombatChart({ analysis }: { analysis: Analysis }) {
  const killed = buildSeries(analysis, "killed");
  const lost = buildSeries(analysis, "lost");
  const net = buildSeries(analysis, "net_value");

  if (killed.length === 0 && lost.length === 0) {
    return <p className="text-gray-500 italic">no combat data</p>;
  }

  const xMax = Math.max(
    Math.ceil(analysis.game.duration_ms / 1000),
    ...killed.flatMap((s) => s.x),
    ...lost.flatMap((s) => s.x),
  );
  const markers = phaseMarkers(analysis);

  return (
    <div className="space-y-3">
      <Panel title="Убил всего (накопительно)" series={killed} xMax={xMax} yLabel="юнитов" markers={markers} />
      <Panel title="Потерял всего (накопительно)" series={lost} xMax={xMax} yLabel="юнитов" markers={markers} />
      <Panel title="Чистый обмен по стоимости (убил − потерял)" series={net} xMax={xMax} yLabel="ресурсов" markers={markers} />
    </div>
  );
}
