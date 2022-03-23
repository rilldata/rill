import type { Patch } from "immer";
import type { DataModelerActionsDefinition } from "$common/data-modeler-service/DataModelerService";
import type { Notification } from "$common/notifications/NotificationService";
import type { EntityType, StateType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityTypeAndStates } from "$common/data-modeler-state-service/DataModelerStateService";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import type { MetricsActionDefinition } from "$common/metrics/MetricsService";

export interface ServerToClientEvents {
  patch: (entityType: EntityType, stateType: StateType, patches: Array<Patch>) => void;
  initialState: (initialStates: EntityTypeAndStates) => void;
  notification: (notification: Notification) => void;
}

export interface ClientToServerEvents {
  action: <Action extends keyof DataModelerActionsDefinition>(
    action: Action, args: DataModelerActionsDefinition[Action],
    callback: (response: ActionResponse) => void
  ) => void;

  metrics: <Event extends keyof MetricsActionDefinition>(
    event: Event, args: MetricsActionDefinition[Event],
  ) => void;
}
