import type { Player } from "../types";

interface Props {
  player: Player;
  compact?: boolean;
}

export function PlayerChip({ player, compact }: Props) {
  const bg = `rgb(${player.color.r}, ${player.color.g}, ${player.color.b})`;
  return (
    <span
      className="inline-block px-2 py-0.5 rounded-full text-white text-xs font-medium"
      style={{ backgroundColor: bg, textShadow: "0 1px 1px rgba(0,0,0,.3)" }}
    >
      {player.name} ({player.civ})
      {player.is_ai && ` · AI d${player.ai_diff ?? "?"}`}
      {!compact && player.team >= 0 && ` · team ${player.team}`}
    </span>
  );
}
