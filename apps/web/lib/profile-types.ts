// Account service wire types, snake_case as sent (domain/port.go).

/** GET /api/auth/statistics?id= */
export interface TimerStatistics {
  user_id: string;
  sessions_joined: number;
  cycles_completed: number;
  total_focus_minutes: number;
  /** "0001-01-01T00:00:00Z" for a user who has never studied. */
  last_active: string;
}

export type RewardType = 'FISH' | 'DECORATION' | 'ROD';
export type RewardRarity = 'COMMON' | 'UNCOMMON' | 'RARE' | 'EPIC' | 'LEGENDARY';

export interface RewardSummaryItem {
  name: string;
  rarity: RewardRarity;
  asset_url: string;
  type: RewardType;
  score_value: number;
  /** Times caught. */
  count: number;
}

/** GET /api/auth/rewards?id= */
export interface RewardsSummary {
  total_awards_earned: number;
  total_score: number;
  items: RewardSummaryItem[];
}
