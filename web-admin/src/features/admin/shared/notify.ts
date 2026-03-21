import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

export function notifySuccess(message: string) {
  eventBus.emit("notification", { type: "success", message });
}

export function notifyError(message: string) {
  eventBus.emit("notification", { type: "error", message });
}
