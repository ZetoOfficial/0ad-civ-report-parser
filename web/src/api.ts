import type { Analysis, ReplayListItem } from "./types";

const base = ""; // same-origin; vite proxies /api in dev

export async function listReplays(): Promise<ReplayListItem[]> {
  const r = await fetch(`${base}/api/replays`);
  if (!r.ok) throw new Error(`/api/replays: ${r.status}`);
  return (await r.json()) ?? [];
}

export async function getReplay(matchID: string): Promise<Analysis> {
  const r = await fetch(`${base}/api/replays/${encodeURIComponent(matchID)}`);
  if (!r.ok) throw new Error(`/api/replays/${matchID}: ${r.status}`);
  return r.json();
}
