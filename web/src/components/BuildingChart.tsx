// Per-template building counts over time. Built from `events` (type=construct)
// so it works even without a sandbox-regenerated metadata.json.
//
// Data caveat: 0 A.D. doesn't emit a `destroy` command in commands.txt, so
// what we plot is *cumulative built* not *currently alive*. For buildings the
// difference is usually small (players rarely tear down their own structures).
import { useMemo, useState } from "react";
import type { Analysis, Player } from "../types";
import { Panel, phaseMarkers, type PlayerSeries } from "./charts";

interface ConstructData { template?: string }

// "structures/spart/civil_centre" → "civil_centre". Display-only.
function shortName(template: string): string {
  const parts = template.split("/");
  return parts[parts.length - 1] || template;
}

interface TemplateStats {
  // Key by structure short-name so that `structures/spart/house` and
  // `structures/maur/house` collapse into a single "house" entry. Civ-unique
  // structures (military_colony, ministry, …) naturally stay singular.
  key: string;
  label: string;
  total: number;
  perPlayer: Map<number, { t: number[]; cum: number[] }>;
}

function collect(analysis: Analysis): TemplateStats[] {
  const acc = new Map<string, TemplateStats>();
  for (const ev of analysis.events) {
    if (ev.type !== "construct") continue;
    const data = ev.data as ConstructData | undefined;
    const tmpl = data?.template;
    if (!tmpl) continue;
    const key = shortName(tmpl);
    let row = acc.get(key);
    if (!row) {
      row = { key, label: key, total: 0, perPlayer: new Map() };
      acc.set(key, row);
    }
    row.total++;
    let pp = row.perPlayer.get(ev.player);
    if (!pp) {
      pp = { t: [], cum: [] };
      row.perPlayer.set(ev.player, pp);
    }
    pp.t.push(ev.t / 1000); // ms → sec
    pp.cum.push(pp.cum.length + 1);
  }
  return Array.from(acc.values()).sort((a, b) => b.total - a.total);
}

function buildSeries(
  stats: TemplateStats,
  playerByID: Record<number, Player>,
  xMax: number,
): PlayerSeries[] {
  const out: PlayerSeries[] = [];
  for (const [pid, pp] of stats.perPlayer) {
    const player = playerByID[pid];
    if (!player) continue;
    // Anchor the line: start at 0 at t=0 and extend the last value out to
    // xMax so the step plot stays flat after the final build.
    const x = [0, ...pp.t, xMax];
    const y = [0, ...pp.cum, pp.cum[pp.cum.length - 1]];
    out.push({ player, x, y });
  }
  return out;
}

export function BuildingChart({ analysis }: { analysis: Analysis }) {
  const stats = useMemo(() => collect(analysis), [analysis]);
  const [selected, setSelected] = useState<Set<string>>(new Set());

  const playerByID = useMemo(() => {
    const m: Record<number, Player> = {};
    for (const p of analysis.players) m[p.id] = p;
    return m;
  }, [analysis]);

  const xMax = Math.ceil(analysis.game.duration_ms / 1000);
  const markers = phaseMarkers(analysis);

  if (stats.length === 0) {
    return <p className="text-gray-500 italic">no construct events</p>;
  }

  function toggle(t: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(t)) next.delete(t);
      else next.add(t);
      return next;
    });
  }

  const visible = stats.filter((s) => selected.has(s.key));

  return (
    <div className="space-y-3">
      <div className="bg-gray-50 rounded-md p-3 border border-gray-100">
        <div className="flex items-baseline justify-between mb-2">
          <h3 className="text-sm font-semibold text-gray-700">
            Здания ({stats.length} типов, {stats.reduce((s, x) => s + x.total, 0)} построек)
          </h3>
          <div className="text-xs text-gray-500 space-x-3">
            <button
              type="button"
              onClick={() => setSelected(new Set(stats.map((s) => s.key)))}
              className="underline hover:text-gray-800"
            >
              все
            </button>
            <button
              type="button"
              onClick={() => setSelected(new Set())}
              className="underline hover:text-gray-800"
            >
              сбросить
            </button>
          </div>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-x-3 gap-y-1 text-sm">
          {stats.map((s) => (
            <label key={s.key} className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={selected.has(s.key)}
                onChange={() => toggle(s.key)}
              />
              <span className="font-mono text-xs">{s.label}</span>
              <span className="text-gray-400 text-xs">·{s.total}</span>
            </label>
          ))}
        </div>
      </div>

      {visible.length === 0 && (
        <p className="text-gray-500 italic text-sm">
          Отметь типы зданий выше — построится по графику на каждый.
        </p>
      )}

      {visible.map((s) => (
        <Panel
          key={s.key}
          title={`${s.label} — построено накопительно`}
          series={buildSeries(s, playerByID, xMax)}
          xMax={xMax}
          yLabel="шт"
          shape="hv"
          markers={markers}
        />
      ))}
    </div>
  );
}
