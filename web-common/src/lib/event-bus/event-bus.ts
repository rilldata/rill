import type {
  BannerEvent,
  PageContentResized,
  NotificationMessage,
} from "./events";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

export interface Events {
  notification: NotificationMessage;
  "clear-all-notifications": void;
  "add-banner": BannerEvent;
  "remove-banner": string;
  "shift-click": void;
  "command-click": void;
  click: void;
  "shift-command-click": void;
  "page-content-resized": PageContentResized;
  "start-chat": string;
  "rill-yaml-updated": void;
  "env-file-updated": string;
  // A project file was written or deleted; the payload is its path, such as `/public/logo.svg`.
  "project-file-changed": string;
  "remote-changes-detected": void;
}

export const eventBus = new EventEmitter<Events>();
