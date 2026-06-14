// Shared X-axis range across every Panel on the page (Grafana-style time
// sync). Selecting a range on any chart updates context → every other Panel
// relayouts to the same window.
import { createContext, useContext, useMemo, useState, type ReactNode } from "react";

export type XRangeSec = [number, number] | null; // seconds; null = full range

interface TimeRangeCtx {
  xRange: XRangeSec;
  setXRange: (r: XRangeSec) => void;
}

const TimeRangeContext = createContext<TimeRangeCtx | null>(null);

export function TimeRangeProvider({ children }: { children: ReactNode }) {
  const [xRange, setXRange] = useState<XRangeSec>(null);
  const value = useMemo(() => ({ xRange, setXRange }), [xRange]);
  return <TimeRangeContext.Provider value={value}>{children}</TimeRangeContext.Provider>;
}

export function useTimeRange(): TimeRangeCtx {
  const c = useContext(TimeRangeContext);
  if (!c) throw new Error("useTimeRange called outside TimeRangeProvider");
  return c;
}
