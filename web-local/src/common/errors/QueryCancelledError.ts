import { ActionErrorType } from "../data-modeler-service/response/ActionResponseMessage";
import type { TypedError } from "./TypedError";

export class QueryCancelledError extends Error implements TypedError {
  public readonly errorType = ActionErrorType.QueryCancelled;

  constructor() {
    super("Query canceled");
  }
}
