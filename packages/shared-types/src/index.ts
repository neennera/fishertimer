// Shared API Contracts and Domain Models for Fisher Timer (v2 3NF)

// ----------------------------------------------------------------------------
// 1. Account Service
// ----------------------------------------------------------------------------
export type UserRole = 'CUSTOMER' | 'ADMIN';

export interface User {
  userId: string;
  email: string;
  displayName: string;
  avatarUrl?: string;
  role: UserRole;
  createdAt: string;
  updatedAt: string;
}

export interface UserStatistics {
  totalSessions: number;
  totalFocusMinutes: number;
  rewardsEarned: number;
}

// ----------------------------------------------------------------------------
// 2. Study Session Service
// ----------------------------------------------------------------------------
export interface StudySession {
  sessionId: string;
  title: string;
  hostId: string;
  isActive: boolean;
  maxParticipants: number;
  createdAt: string;
  endedAt?: string;
}

export interface SessionParticipant {
  sessionId: string;
  userId: string;
  joinedAt: string;
  leftAt?: string;
}

// ----------------------------------------------------------------------------
// 3. Study Timer Service
// ----------------------------------------------------------------------------
export interface TimerSettings {
  userId: string;
  focusDuration: number;
  shortBreakDuration: number;
  longBreakDuration: number;
  cyclesBeforeLongBreak: number;
  updatedAt: string;
}

export type TimerPhaseType = 'FOCUS' | 'SHORT_BREAK' | 'LONG_BREAK';
export type TimerSessionStatus = 'FOCUS' | 'SHORT_BREAK' | 'LONG_BREAK' | 'PAUSED' | 'COMPLETED' | 'STOPPED';

export interface TimerSession {
  timerSessionId: string;
  userId: string;
  studySessionId?: string;
  status: TimerSessionStatus;
  startedAt: string;
  completedAt?: string;
}

export interface TimerCycle {
  cycleId: string;
  timerSessionId: string;
  cycleNumber: number;
  phaseType: TimerPhaseType;
  duration: number;
  isCompleted: boolean;
  endedAt: string;
}

// ----------------------------------------------------------------------------
// 4. Reward Service
// ----------------------------------------------------------------------------
export type RewardItemType = 'SKIN' | 'BADGE' | 'FISH_SPECIES';

export interface RewardItem {
  id: string;
  itemName: string;
  description?: string;
  itemType: RewardItemType;
  cost: number;
  createdAt: string;
}

export interface UserReward {
  id: string;
  userId: string;
  itemId: string;
  unlockedAt: string;
}

// ----------------------------------------------------------------------------
// 5. Admin Service
// ----------------------------------------------------------------------------
export type AdminAction = 'FORCE_CLOSE_SESSION' | 'KICK_USER';

export interface AdminLog {
  logId: string;
  adminId: string;
  action: AdminAction;
  targetId: string;
  reason?: string;
  createdAt: string;
}

// ----------------------------------------------------------------------------
// 6. Leaderboard Service
// ----------------------------------------------------------------------------
export interface LeaderboardEntry {
  userId: string;
  displayName: string;
  rank: number;
  score: number;
}
