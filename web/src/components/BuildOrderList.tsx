import type { Analysis, ReplayEvent } from "../types";

const KEEP_STRUCTURE_SUBSTRINGS = [
  "wonder", "civic_centre", "fortress",
  "barracks", "stable", "elephant_stable", "kennel",
  "range", "embassy", "temple", "workshop",
  "tower", "wallset", "dock", "market",
  "library", "archive", "pyramid",
];

const SIEGE_SUBSTRINGS = [
  "catapult", "lithobolos", "oxybeles", "scorpio",
  "ram", "siege_",
];

function eventLabel(
  e: ReplayEvent,
  seenChampion: Set<number>,
  seenSiege: Set<number>,
): string | null {
  if (e.type === "resign") return "RESIGN";

  const data = e.data as { template?: string } | undefined;
  const tmpl = data?.template ?? "";

  if (e.type === "research") {
    if (tmpl.startsWith("phase_")) return `research ${tmpl}`;
    return null;
  }

  if (e.type === "construct") {
    if (KEEP_STRUCTURE_SUBSTRINGS.some((k) => tmpl.includes(k))) {
      return `construct ${tmpl}`;
    }
    return null;
  }

  if (e.type === "train") {
    if (tmpl.includes("/hero_")) return `train hero ${tmpl}`;
    if (tmpl.includes("champion") && !seenChampion.has(e.player)) {
      seenChampion.add(e.player);
      return `train first champion ${tmpl}`;
    }
    if (SIEGE_SUBSTRINGS.some((k) => tmpl.includes(k)) && !seenSiege.has(e.player)) {
      seenSiege.add(e.player);
      return `train first siege ${tmpl}`;
    }
    return null;
  }

  return null;
}

function fmtTime(ms: number): string {
  const m = Math.floor(ms / 60000);
  const s = Math.floor((ms % 60000) / 1000);
  return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
}

interface Props { analysis: Analysis }

export function BuildOrderList({ analysis }: Props) {
  const playerName: Record<number, string> = {};
  for (const p of analysis.players) playerName[p.id] = p.name;

  const seenChampion = new Set<number>();
  const seenSiege = new Set<number>();
  const rows: { time: string; player: string; event: string }[] = [];

  for (const e of analysis.events) {
    const label = eventLabel(e, seenChampion, seenSiege);
    if (!label) continue;
    rows.push({
      time: fmtTime(e.t),
      player: playerName[e.player] ?? `p${e.player}`,
      event: label,
    });
  }

  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="text-left text-gray-500">
          <th className="font-normal py-1">t</th>
          <th className="font-normal py-1">player</th>
          <th className="font-normal py-1">event</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((r, i) => (
          <tr key={i} className="border-t border-gray-100">
            <td className="py-1 pr-2 font-mono">{r.time}</td>
            <td className="py-1 pr-2">{r.player}</td>
            <td className="py-1">{r.event}</td>
          </tr>
        ))}
        {rows.length === 0 && (
          <tr><td colSpan={3} className="text-gray-500 py-2">no significant events</td></tr>
        )}
      </tbody>
    </table>
  );
}
