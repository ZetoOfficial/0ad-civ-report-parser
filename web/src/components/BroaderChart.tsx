// Buildings lost / map control / non-gather income. Each line per player.
import type { Analysis } from "../types";
import { Panel, phaseMarkers, type PlayerSeries } from "./charts";

type Kind = "buildings_lost" | "map_controlled" | "trade_income" | "loot";

function buildSeries(analysis: Analysis, kind: Kind): PlayerSeries[] {
  const out: PlayerSeries[] = [];
  for (const p of analysis.players) {
    const pm = analysis.metrics.players[String(p.id)];
    const s = pm?.sequences;
    if (!s || s.time.length === 0) continue;
    let y: number[] = [];
    if (kind === "buildings_lost") {
      // 0 A.D. uses "Structure" (catch-all) or "total" (lowercase) — try both.
      y = s.buildings_lost.Structure ?? s.buildings_lost.total ?? [];
    } else if (kind === "map_controlled") {
      y = s.percent_map_controlled;
    } else if (kind === "trade_income") {
      y = s.trade_income;
    } else if (kind === "loot") {
      y = s.loot_collected;
    }
    if (!y || y.length === 0) continue;
    out.push({ player: p, x: s.time, y });
  }
  return out;
}

export function BroaderChart({ analysis }: { analysis: Analysis }) {
  const bldg = buildSeries(analysis, "buildings_lost");
  const map = buildSeries(analysis, "map_controlled");
  const trade = buildSeries(analysis, "trade_income");
  const loot = buildSeries(analysis, "loot");

  const xMax = Math.max(
    Math.ceil(analysis.game.duration_ms / 1000),
    ...map.flatMap((s) => s.x),
  );
  const markers = phaseMarkers(analysis);

  return (
    <div className="space-y-3">
      <Panel title="Потерянные здания (накопительно)" series={bldg} xMax={xMax} yLabel="зданий" markers={markers} shape="hv" />
      <Panel title="Контролируемая карта" series={map} xMax={xMax} yLabel="%" markers={markers} />
      <Panel title="Доход с торговли (накопительно)" series={trade} xMax={xMax} yLabel="ресурсов" markers={markers} />
      <Panel title="Собранный лут (накопительно)" series={loot} xMax={xMax} yLabel="ресурсов" markers={markers} />
    </div>
  );
}
