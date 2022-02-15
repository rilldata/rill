export interface Notification {
    message: string;
    type: string;
}

export interface NotificationService {
    notify(notification: Notification): void
}
