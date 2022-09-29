import type { Patch } from "immer";
import type { DataModelerActionsDefinition } from "../data-modeler-service/DataModelerService";
import type { Notification } from "../notifications/NotificationService";
import type {
  EntityType,
  StateType,
} from "../data-modeler-state-service/entity-state-service/EntityStateService";
import type { EntityTypeAndStates } from "../data-modeler-state-service/DataModelerStateService";
import type { ActionResponse } from "../data-modeler-service/response/ActionResponse";
import type { MetricsActionDefinition } from "../metrics-service/MetricsService";

export interface ServerToClientEvents {
  patch: (
    entityType: EntityType,
    stateType: StateType,
    patches: Array<Patch>
  ) => void;
  initialState: (initialStates: EntityTypeAndStates) => void;
  notification: (notification: Notification) => void;
}

export interface ClientToServerEvents {
  action: <Action extends keyof DataModelerActionsDefinition>(
    action: Action,
    args: DataModelerActionsDefinition[Action],
    callback: (response: ActionResponse) => void
  ) => void;

  metrics: <Event extends keyof MetricsActionDefinition>(
    event: Event,
    args: MetricsActionDefinition[Event]
  ) => void;
}
