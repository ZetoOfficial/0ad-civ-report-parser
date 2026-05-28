import type { Analysis } from "../types";

interface Props { analysis: Analysis }

export function AnomalyList({ analysis }: Props) {
  const rows: { type: string; severity: string; t: string; detail: string }[] = [];
  for (const m of Object.values(analysis.metrics.players)) {
    for (const an of m.anomalies ?? []) {
      let detail = "";
      const d = an.details as Record<string, unknown> | undefined;
      if (d?.target !== undefined) detail = `target=${d.target}`;
      rows.push({
        type: an.type,
        severity: an.severity,
        t: `${an.t_start_sec}..${an.t_end_sec}s`,
        detail,
      });
    }
  }

  if (rows.length === 0) return <p className="text-gray-500 italic">Чисто</p>;

  return (
    <ul className="space-y-1">
      {rows.map((r, i) => (
        <li key={i}>
          <span className="font-semibold">{r.type}</span>
          <span className="text-gray-500"> · {r.severity} · {r.t} · {r.detail}</span>
        </li>
      ))}
    </ul>
  );
}
