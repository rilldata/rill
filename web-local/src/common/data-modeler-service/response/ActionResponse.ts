import type { ActionResponseMessage } from "./ActionResponseMessage";

export enum ActionStatus {
  Success,
  Failure,
}

export interface ActionResponse {
  status: ActionStatus;
  messages: Array<ActionResponseMessage>;
  data?: unknown;
}
