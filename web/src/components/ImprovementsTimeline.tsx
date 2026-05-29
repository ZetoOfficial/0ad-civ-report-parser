import { useMemo } from "react";
import type { Analysis, Improvement } from "../types";
import { PlayerChip } from "./PlayerChip";
import { formatDuration } from "../utils";

interface Props {
  analysis: Analysis;
}

function formatCost(imp: Improvement): string {
  if (!imp.cost) return "";
  const parts: string[] = [];
  if (imp.cost.food) parts.push(`F:${imp.cost.food}`);
  if (imp.cost.wood) parts.push(`W:${imp.cost.wood}`);
  if (imp.cost.stone) parts.push(`S:${imp.cost.stone}`);
  if (imp.cost.metal) parts.push(`M:${imp.cost.metal}`);
  if (parts.length === 0) return "";
  return `(${parts.join(" ")})`;
}

interface BuildingGroup {
  building: string;
  entries: Improvement[];
}

function groupByBuilding(improvements: Improvement[]): BuildingGroup[] {
  const acc = new Map<string, Improvement[]>();
  for (const imp of improvements) {
    const key = imp.building || "—";
    const arr = acc.get(key);
    if (arr) {
      arr.push(imp);
    } else {
      acc.set(key, [imp]);
    }
  }
  // Sort by total count descending.
  const groups: BuildingGroup[] = [];
  for (const [building, entries] of acc) {
    // Sort entries within group by t_ms ascending.
    entries.sort((a, b) => a.t_ms - b.t_ms);
    groups.push({ building, entries });
  }
  groups.sort((a, b) => b.entries.length - a.entries.length);
  return groups;
}

export function ImprovementsTimeline({ analysis }: Props) {
  const playerByID = useMemo(() => {
    const m: Record<number, Analysis["players"][number]> = {};
    for (const p of analysis.players) m[p.id] = p;
    return m;
  }, [analysis]);

  return (
    <div className="space-y-6">
      {analysis.players.map((player) => {
        const raw = analysis.metrics.players[String(player.id)]?.improvements ?? [];
        // Filter out auto-research entries (civ bonuses, not player decisions).
        const improvements = raw.filter((imp) => !imp.auto_research);

        const groups = groupByBuilding(improvements);

        return (
          <div key={player.id}>
            <div className="mb-2">
              <PlayerChip player={player} />
            </div>
            {improvements.length === 0 ? (
              <p className="text-gray-500 italic text-sm">нет исследований</p>
            ) : (
              <div className="space-y-3">
                {groups.map((group) => (
                  <div
                    key={group.building}
                    className="bg-gray-50 rounded-md p-3 border border-gray-100"
                  >
                    <h3 className="text-sm font-semibold text-gray-700 mb-2 font-mono">
                      {group.building}
                    </h3>
                    <ul className="space-y-1">
                      {group.entries.map((imp, idx) => {
                        const label = imp.generic_name || imp.template;
                        const cost = formatCost(imp);
                        return (
                          <li
                            key={idx}
                            className="flex items-baseline gap-2 text-sm"
                            title={imp.description ?? ""}
                          >
                            <span className="font-mono text-xs text-gray-400 shrink-0">
                              [{formatDuration(imp.t_ms)}]
                            </span>
                            <span className="text-gray-800">{label}</span>
                            {cost && (
                              <span className="text-gray-400 text-xs">{cost}</span>
                            )}
                          </li>
                        );
                      })}
                    </ul>
                  </div>
                ))}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
