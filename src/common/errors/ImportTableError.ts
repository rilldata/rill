import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";
import type { TypedError } from "$common/errors/TypedError";

export class ImportTableError extends Error implements TypedError {
  public readonly errorType = ActionErrorType.ImportTable;
}
