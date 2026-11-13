// HTTP notification implementation — identical interface to the mock (api wave).

import { apiClient } from '../../lib/apiClient';
import type { NotificationService } from './index';
import type { InboxMessage, Notification, NotificationSettings } from '../../types';
import type { Paginated } from '../../types/common';
import {
  toInboxMessage,
  toNotification,
  type ApiInboxMessage,
  type ApiNotification,
} from '../apiMappers';

export const httpNotificationService: NotificationService = {
  async list(page = 1, pageSize = 20): Promise<Paginated<Notification>> {
    const { data, meta } = await apiClient.get<ApiNotification[]>('/notifications', {
      page,
      pageSize,
    });
    const items = (data ?? []).map(toNotification);
    return {
      items,
      meta: {
        page: meta?.page ?? page,
        pageSize: meta?.pageSize ?? pageSize,
        total: meta?.total ?? items.length,
        hasMore: meta?.hasMore ?? false,
      },
    };
  },

  async markRead(id: number): Promise<void> {
    await apiClient.post(`/notifications/${id}/read`);
  },

  async inbox(): Promise<InboxMessage[]> {
    const { data } = await apiClient.get<ApiInboxMessage[]>('/inbox');
    return (data ?? []).map(toInboxMessage);
  },

  async updateSettings(settings: NotificationSettings): Promise<NotificationSettings> {
    const { data } = await apiClient.put<NotificationSettings>(
      '/profile/notification-settings',
      settings,
    );
    return data;
  },
};
