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

export interface DensityBin {
  t_sec: number;
  counts: Record<string, number>;
}

export interface PlayerMetrics {
  phase_timings: Record<string, number>;
  engagements: Engagement[];
  anomalies: Anomaly[];
}

export interface Metrics {
  players: Record<string, PlayerMetrics>;
  action_density: DensityBin[];
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
