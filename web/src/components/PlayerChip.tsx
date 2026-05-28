import type { Player } from "../types";
import { colorToCss, contrastText } from "../utils";

interface Props {
  player: Player;
  compact?: boolean;
}

export function PlayerChip({ player, compact }: Props) {
  const bg = colorToCss(player.color);
  const fg = contrastText(player.color);
  return (
    <span
      className="inline-block px-2 py-0.5 rounded-full text-xs font-medium"
      style={{ backgroundColor: bg, color: fg }}
    >
      {player.name} ({player.civ})
      {player.is_ai && ` · AI d${player.ai_diff ?? "?"}`}
      {!compact && player.team >= 0 && ` · team ${player.team}`}
    </span>
  );
}
