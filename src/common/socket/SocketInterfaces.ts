import type { Patch } from "immer";
import type { DataModelerState } from "$lib/types";
import type { DataModelerActionsDefinition } from "$common/data-modeler-service/DataModelerService";
import type { Notification } from "$common/notifications/NotificationService";

export interface ServerToClientEvents {
  patch: (patches: Array<Patch>) => void;
  initialState: (initialState: DataModelerState) => void;
  notification: (notification: Notification) => void;
}

export interface ClientToServerEvents {
  action: <Action extends keyof DataModelerActionsDefinition>(
    action: Action, args: DataModelerActionsDefinition[Action]
  ) => void;
}
