import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listReplays } from "../api";
import type { ReplayListItem } from "../types";
import { PlayerChip } from "../components/PlayerChip";
import { formatDuration } from "../utils";

function formatDate(unix: number): string {
  if (!unix) return "—";
  return new Date(unix * 1000).toLocaleString("ru-RU", {
    day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit",
  });
}

export function IndexPage() {
  const [items, setItems] = useState<ReplayListItem[] | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    const ac = new AbortController();
    listReplays(ac.signal)
      .then((x) => { if (!ac.signal.aborted) setItems(x); })
      .catch((e) => { if (!ac.signal.aborted) setErr(String(e)); });
    return () => ac.abort();
  }, []);

  if (err) return <p className="text-red-600">Error: {err}</p>;
  if (items === null) return <p>Loading…</p>;
  if (items.length === 0) return <p>No replays yet.</p>;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Партии ({items.length})</h1>
      <ul className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
        {items.map((it) => (
          <li key={it.dir} className="bg-white rounded-lg border border-gray-200 hover:border-gray-400 transition">
            <Link to={`/replay/${it.match_id}`} className="block p-4">
              <div className="font-semibold">{it.map || "—"}</div>
              <div className="text-xs text-gray-500 mt-1">
                {formatDate(it.timestamp)} · {formatDuration(it.duration_ms)}
              </div>
              <div className="flex flex-wrap gap-1 my-2">
                {it.players.map((p) => <PlayerChip key={p.id} player={p} compact />)}
              </div>
              <div className="text-xs text-gray-600">{it.outcome}</div>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
