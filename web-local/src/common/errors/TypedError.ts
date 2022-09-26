import type { ActionErrorType } from "../data-modeler-service/response/ActionResponseMessage";

export interface TypedError {
  errorType?: ActionErrorType;
}
