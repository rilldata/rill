import type { Readable } from "svelte/store";
import { derived, writable } from "svelte/store";

const NOTIFICATION_TIMEOUT = 2000;

interface NotificationStore extends Readable<object> {
  timeoutID: ReturnType<typeof setTimeout>;
  send: (args: NotificationMessageArguments) => void;
  clear: () => void;
}

interface NotificationMessageArguments {
  message: string;
  type?: string;
  detail?: string;
  options?: NotificationOptions;
}

interface NotificationMessage {
  id: string;
  type?: string;
  message: string;
  detail?: string;
  options?: NotificationOptions;
}

// No need to export after we deprecate the Node backend
export interface NotificationOptions {
  width?: number;
  persisted?: boolean;
}

function createNotificationStore(): NotificationStore {
  const _notification = writable({} as NotificationMessage);
  let timeout: ReturnType<typeof setTimeout>;

  function send({
    message,
    type = "default",
    detail,
    options = {},
  }: NotificationMessageArguments): void {
    const notificationMessage: NotificationMessage = {
      id: id(),
      message,
      type,
      detail,
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
      if ($notification.id && !$notification.options?.persisted) {
        timeout = setTimeout(clear, NOTIFICATION_TIMEOUT);
      }
    }
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
