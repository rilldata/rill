import { ActionErrorType } from "../data-modeler-service/response/ActionResponseMessage";
import type { TypedError } from "./TypedError";

export class ModelQueryError extends Error implements TypedError {
  public readonly errorType = ActionErrorType.ModelQuery;
}
