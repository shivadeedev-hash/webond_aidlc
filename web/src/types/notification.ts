// Notification + simulated-email inbox DTOs (data-model.md: notification_outbox).

import type { ISODateTime } from './common';

export type NotificationChannel = 'inapp' | 'email';

export type NotificationCategory =
  | 'verify'
  | 'reset'
  | 'order'
  | 'dca'
  | 'coupon'
  | 'maturity'
  | 'general';

export interface Notification {
  id: number;
  channel: NotificationChannel;
  category: NotificationCategory;
  subject: string;
  body: string;
  readAt: ISODateTime | null;
  createdAt: ISODateTime;
}

/** Simulated email inbox item (FR-26) — may carry an OTP/reset token for dev. */
export interface InboxMessage extends Notification {
  channel: 'email';
  token: string | null;
  tokenExpiresAt: ISODateTime | null;
}

export interface NotificationSettings {
  emailEnabled: boolean;
}
