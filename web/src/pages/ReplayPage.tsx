import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getReplay } from "../api";
import type { Analysis } from "../types";
import { PlayerChip } from "../components/PlayerChip";
import { BuildOrderList } from "../components/BuildOrderList";
import { AnomalyList } from "../components/AnomalyList";
import { DensityChart } from "../components/DensityChart";

export function ReplayPage() {
  const { matchID = "" } = useParams();
  const [a, setA] = useState<Analysis | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    getReplay(matchID).then(setA).catch((e) => setErr(String(e)));
  }, [matchID]);

  if (err) return <p className="text-red-600">Error: {err}</p>;
  if (!a) return <p>Loading…</p>;

  return (
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold">{a.game.map}</h1>
        <div className="text-sm text-gray-500 flex gap-4 mt-1">
          <span>matchID: {a.game.match_id}</span>
          <span>длительность: {Math.floor(a.game.duration_ms / 60000)} мин</span>
          <span>движок: {a.game.engine_version}</span>
        </div>
      </header>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Игроки</h2>
        <div className="space-y-2">
          {a.players.map((p) => {
            const fs = a.final_state.players[String(p.id)];
            return (
              <div key={p.id} className="flex items-center gap-3 text-sm">
                <PlayerChip player={p} />
                {fs && (
                  <>
                    <span className="font-semibold">{fs.outcome}</span>
                    <span className="text-gray-500">фаза: {fs.phase || "—"}</span>
                    <span className="text-gray-500">поп: {fs.pop_count}/{fs.pop_limit}</span>
                  </>
                )}
              </div>
            );
          })}
        </div>
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Плотность действий (30 сек)</h2>
        <DensityChart analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Build order (значимые события)</h2>
        <BuildOrderList analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Аномалии</h2>
        <AnomalyList analysis={a} />
      </section>
    </div>
  );
}
