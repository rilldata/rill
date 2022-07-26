import type { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export interface TypedError {
  errorType: ActionErrorType;
}
