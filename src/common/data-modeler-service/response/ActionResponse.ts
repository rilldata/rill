import type { ActionResponseMessage } from "$common/data-modeler-service/response/ActionResponseMessage";

export enum ActionStatus {
  Success,
  Failure,
}

export interface ActionResponse {
  status: ActionStatus;
  messages: Array<ActionResponseMessage>;
}
