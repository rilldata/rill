import type { Readable } from "svelte/store";
import { derived, writable } from "svelte/store";

const NOTIFICATION_TIMEOUT = 2000;

interface NotificationStore extends Readable<NotificationMessage> {
  timeoutID: ReturnType<typeof setTimeout>;
  send: (args: NotificationMessageArguments) => void;
  clear: () => void;
}

export interface Link {
  text: string;
  href: string;
}

interface NotificationMessageArguments {
  message: string;
  type?: string;
  detail?: string;
  link?: Link;
  options?: NotificationOptions;
}

interface NotificationMessage {
  id: string;
  type?: string;
  message: string;
  detail?: string;
  link?: Link;
  options?: NotificationOptions;
}

interface NotificationOptions {
  width?: number;
  persisted?: boolean;
  persistedLink?: boolean;
}

function createNotificationStore(): NotificationStore {
  const _notification = writable({} as NotificationMessage);
  let timeout: ReturnType<typeof setTimeout>;

  function send({
    message,
    type = "default",
    detail,
    link,
    options = {},
  }: NotificationMessageArguments): void {
    const notificationMessage: NotificationMessage = {
      id: id(),
      message,
      type,
      detail,
      link,
      options,
    };
    _notification.set(notificationMessage);
  }

  function clear(): void {
    _notification.set({} as NotificationMessage);
  }

  const notifications: Readable<object> = derived(
    _notification,
    ($notification, set) => {
      // if there already was a notification, let's clear the timer
      // and reset it here.
      clearTimeout(timeout);
      set($notification);
      // if this is not the reset message, set the timer.
      if (
        $notification.id &&
        !$notification.options?.persisted &&
        !$notification.options?.persistedLink
      ) {
        timeout = setTimeout(clear, NOTIFICATION_TIMEOUT);
      }
    },
  );
  const { subscribe } = notifications;

  return {
    timeoutID: timeout,
    subscribe,
    send,
    clear: () => {
      clearTimeout(timeout);
      clear();
    },
  };
}

function id(): string {
  return "_" + Math.random().toString(36).substr(2, 9);
}

export default createNotificationStore();
