// Mock notification implementation backed by notifications.json.

import notificationData from '../../mocks/notifications.json';
import { mockResolve, paginate } from '../mockUtils';
import type { NotificationService } from './index';
import type { InboxMessage, Notification, NotificationSettings } from '../../types';

const notifications = structuredClone(notificationData.notifications) as Notification[];
const inboxMessages = notificationData.inbox as InboxMessage[];
let emailEnabled = true;

export const mockNotificationService: NotificationService = {
  async list(page = 1, pageSize = 20) {
    const sorted = [...notifications].sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
    );
    return mockResolve(paginate(sorted, page, pageSize));
  },

  async markRead(id: number): Promise<void> {
    const n = notifications.find((x) => x.id === id);
    if (n && !n.readAt) n.readAt = new Date().toISOString();
    await mockResolve(null);
  },

  async inbox(): Promise<InboxMessage[]> {
    return mockResolve(inboxMessages);
  },

  async updateSettings(settings: NotificationSettings): Promise<NotificationSettings> {
    emailEnabled = settings.emailEnabled;
    return mockResolve({ emailEnabled });
  },
};
