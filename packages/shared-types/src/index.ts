// Shared API Contracts and Domain Models for Fisher Timer

export interface User {
  id: string;
  email: string;
  displayName: string;
  isBanned: boolean;
  createdAt: string;
}

export interface StudySession {
  id: string;
  name: string;
  creatorId: string;
  participantLimit: number;
  status: 'ACTIVE' | 'ENDED';
  createdAt: string;
}

export interface SessionParticipant {
  sessionId: string;
  userId: string;
  joinedAt: string;
  leftAt?: string;
}

export type TimerStatus = 'STOPPED' | 'RUNNING' | 'PAUSED' | 'RESTING';

export interface TimerSession {
  sessionId: string;
  userId: string;
  status: TimerStatus;
  workDurationMinutes: number;
  restDurationMinutes: number;
  currentCycle: number;
}

export interface Reward {
  id: string;
  userId: string;
  name: string;
  rarity: 'COMMON' | 'UNCOMMON' | 'RARE' | 'EPIC' | 'LEGENDARY';
  awardedAt: string;
}

export interface LeaderboardEntry {
  userId: string;
  displayName: string;
  rank: number;
  totalRewardCount: number;
  totalFocusMinutes: number;
}
