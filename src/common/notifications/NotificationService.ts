import type { NotificationOptions } from "$lib/components/notifications";

export interface Notification {
  message: string;
  type: string;
  detail?: string;
  options?: NotificationOptions;
}

export interface NotificationService {
  notify(notification: Notification): void;
}
