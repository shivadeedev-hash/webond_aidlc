// Notification service interface + env-selected implementation
// (api-spec.md /api/notifications, /api/inbox, /api/profile/notification-settings).

import { DATA_SOURCE } from '../../lib/constants';
import type { InboxMessage, Notification, NotificationSettings } from '../../types';
import type { Paginated } from '../../types/common';
import { mockNotificationService } from './mock';
import { httpNotificationService } from './http';

export interface NotificationService {
  list(page?: number, pageSize?: number): Promise<Paginated<Notification>>;
  markRead(id: number): Promise<void>;
  inbox(): Promise<InboxMessage[]>;
  updateSettings(settings: NotificationSettings): Promise<NotificationSettings>;
}

export const notificationService: NotificationService =
  DATA_SOURCE === 'api' ? httpNotificationService : mockNotificationService;
