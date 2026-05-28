import type { Analysis } from "../types";
import { PlayerChip } from "./PlayerChip";

interface Props { analysis: Analysis }

export function AnomalyList({ analysis }: Props) {
  const rows: { pid: number; type: string; severity: string; t: string; detail: string }[] = [];
  for (const [pidStr, m] of Object.entries(analysis.metrics.players)) {
    const pid = Number(pidStr);
    for (const an of m.anomalies ?? []) {
      let detail = "";
      const d = an.details as Record<string, unknown> | undefined;
      if (d?.target !== undefined) detail = `target=${String(d.target)}`;
      rows.push({
        pid,
        type: an.type,
        severity: an.severity,
        t: `${an.t_start_sec}..${an.t_end_sec}s`,
        detail,
      });
    }
  }
  rows.sort((a, b) => a.pid - b.pid);

  if (rows.length === 0) return <p className="text-gray-500 italic">Чисто</p>;

  const playerByID: Record<number, Analysis["players"][number]> = {};
  for (const p of analysis.players) playerByID[p.id] = p;

  return (
    <ul className="space-y-1">
      {rows.map((r, i) => {
        const p = playerByID[r.pid];
        return (
          <li key={i} className="flex items-center gap-2">
            {p ? <PlayerChip player={p} compact /> : <span className="text-xs">p{r.pid}</span>}
            <span className="font-semibold">{r.type}</span>
            <span className="text-gray-500 text-sm">· {r.severity} · {r.t}{r.detail && ` · ${r.detail}`}</span>
          </li>
        );
      })}
    </ul>
  );
}
