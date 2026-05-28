export interface Color {
  r: number;
  g: number;
  b: number;
}

export interface Player {
  id: number;
  name: string;
  civ: string;
  team: number;
  is_ai: boolean;
  ai_diff?: number;
  color: Color;
}

export interface Resources {
  food: number;
  wood: number;
  stone: number;
  metal: number;
}

export interface PlayerFinalState {
  phase: string;
  state: string;
  outcome: string;
  pop_count: number;
  pop_limit: number;
  pop_max: number;
  resource_counts: Resources;
  researched_techs: string[];
}

export interface FinalState {
  players: Record<string, PlayerFinalState>; // JSON keys are stringified ints
}

export interface Engagement {
  t_start_sec: number;
  t_end_sec: number;
  target: number;
  peak_units: number;
  command_count: number;
}

export interface Anomaly {
  type: string;
  t_start_sec: number;
  t_end_sec: number;
  severity: string;
  details?: Record<string, unknown>;
}

// ClassSeries: per-class cumulative time-series. Each array has the same
// length as Sequences.time.
export type ClassSeries = Record<string, number[]>;

export interface Sequences {
  time: number[]; // x-axis, seconds from game start (~30s sampling)

  population_count: number[];
  percent_map_explored: number[];
  percent_map_controlled: number[];

  units_trained: ClassSeries;
  units_lost: ClassSeries;
  enemy_units_killed: ClassSeries;
  units_captured: ClassSeries;
  buildings_constructed: ClassSeries;
  buildings_lost: ClassSeries;
  enemy_buildings_destroyed: ClassSeries;
  buildings_captured: ClassSeries;

  units_lost_value: number[];
  enemy_units_killed_value: number[];
  units_captured_value: number[];
  buildings_lost_value: number[];
  enemy_buildings_destroyed_value: number[];
  buildings_captured_value: number[];

  trade_income: number[];
  tributes_sent: number[];
  tributes_received: number[];
  loot_collected: number[];
  treasures_collected: number[];
  successful_bribes: number[];
  failed_bribes: number[];
}

export interface PlayerMetrics {
  phase_timings: Record<string, number>;
  engagements: Engagement[];
  anomalies: Anomaly[];
  sequences?: Sequences | null;
}

export interface Metrics {
  players: Record<string, PlayerMetrics>;
}

export interface GameInfo {
  match_id: string;
  map: string;
  map_type: string;
  timestamp: number;
  duration_ms: number;
  engine_version: string;
  victory_conditions: string[];
}

export interface ReplayEvent {
  t: number;
  player: number;
  type: string;
  data?: unknown;
}

export interface Analysis {
  schema_version: number;
  game: GameInfo;
  players: Player[];
  events: ReplayEvent[];
  snapshots: unknown[];
  final_state: FinalState;
  metrics: Metrics;
}

export interface ReplayListItem {
  dir: string;
  match_id: string;
  map: string;
  timestamp: number;
  duration_ms: number;
  players: Player[];
  outcome: string;
}
