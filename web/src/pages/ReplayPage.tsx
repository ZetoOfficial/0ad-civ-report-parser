import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getReplay } from "../api";
import type { Analysis } from "../types";
import { PlayerChip } from "../components/PlayerChip";
import { AnomalyList } from "../components/AnomalyList";
import { PopulationChart } from "../components/PopulationChart";
import { CombatChart } from "../components/CombatChart";
import { BroaderChart } from "../components/BroaderChart";
import { ResourceChart } from "../components/ResourceChart";
import { BuildingChart } from "../components/BuildingChart";
import { ImprovementsTimeline } from "../components/ImprovementsTimeline";
import { formatDuration } from "../utils";
import { TimeRangeProvider, useTimeRange } from "../timeRange";

function TimeRangeBar() {
  const { xRange, setXRange } = useTimeRange();
  if (!xRange) return null;
  return (
    <div className="bg-yellow-50 border border-yellow-200 rounded-md p-2 flex items-center justify-between text-sm">
      <span>
        Зум: <span className="font-mono">{formatDuration(xRange[0] * 1000)}</span>
        {" — "}
        <span className="font-mono">{formatDuration(xRange[1] * 1000)}</span>
        {" "}<span className="text-gray-500">(синхронизирован на все графики ниже)</span>
      </span>
      <button
        type="button"
        onClick={() => setXRange(null)}
        className="px-2 py-1 rounded bg-white border border-gray-300 hover:border-gray-500"
      >
        Сбросить зум
      </button>
    </div>
  );
}

export function ReplayPage() {
  const { matchID = "" } = useParams();
  const [a, setA] = useState<Analysis | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    setA(null);
    setErr(null);
    const ac = new AbortController();
    getReplay(matchID, ac.signal)
      .then((x) => { if (!ac.signal.aborted) setA(x); })
      .catch((e) => {
        if (ac.signal.aborted) return;
        if (e instanceof DOMException && e.name === "AbortError") return;
        setErr(String(e));
      });
    return () => ac.abort();
  }, [matchID]);

  if (err) return <p className="text-red-600">Error: {err}</p>;
  if (!a) return <p>Loading…</p>;

  return (
    <TimeRangeProvider>
    <div className="space-y-6">
      <header>
        <h1 className="text-2xl font-bold">{a.game.map}</h1>
        <div className="text-sm text-gray-500 flex gap-4 mt-1">
          <span>matchID: {a.game.match_id}</span>
          <span>длительность: {formatDuration(a.game.duration_ms)}</span>
          <span>движок: {a.game.engine_version}</span>
        </div>
      </header>

      <TimeRangeBar />

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
        <h2 className="font-semibold mb-2">Популяция, работяги, армия (живых)</h2>
        <PopulationChart analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Бой: убитые / потерянные / обмен по стоимости</h2>
        <CombatChart analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Ресурсы по типам</h2>
        <ResourceChart analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Здания / карта / эко-добавки</h2>
        <BroaderChart analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Здания по типу (выбери в чеклисте)</h2>
        <BuildingChart analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Улучшения по зданиям</h2>
        <ImprovementsTimeline analysis={a} />
      </section>

      <section className="bg-white p-4 rounded-lg border border-gray-200">
        <h2 className="font-semibold mb-2">Аномалии</h2>
        <AnomalyList analysis={a} />
      </section>
    </div>
    </TimeRangeProvider>
  );
}
