import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listReplays } from "../api";
import type { ReplayListItem } from "../types";
import { PlayerChip } from "../components/PlayerChip";

function formatDate(unix: number): string {
  if (!unix) return "—";
  return new Date(unix * 1000).toLocaleString("ru-RU", {
    day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit",
  });
}

function formatDuration(ms: number): string {
  const min = Math.floor(ms / 60000);
  const sec = Math.floor((ms % 60000) / 1000);
  return `${min}:${sec.toString().padStart(2, "0")}`;
}

export function IndexPage() {
  const [items, setItems] = useState<ReplayListItem[] | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    listReplays().then(setItems).catch((e) => setErr(String(e)));
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
