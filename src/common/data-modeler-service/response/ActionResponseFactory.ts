import { ActionResponseMessageType } from "$common/data-modeler-service/response/ActionResponseMessage";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import { ImportTableError } from "$common/errors/ImportTableError";
import { ModelQueryError } from "$common/errors/ModelQueryError";
import { EntityError } from "$common/errors/EntityError";

export class ActionResponseFactory {
  public static getSuccessResponse(message?: string): ActionResponse {
    return {
      messages: message
        ? [
            {
              type: ActionResponseMessageType.Info,
              message,
            },
          ]
        : [],
      status: ActionStatus.Success,
    };
  }

  public static getErrorResponse(error: Error): ActionResponse {
    return {
      messages: [
        {
          type: ActionResponseMessageType.Error,
          message: error.message,
          stack: error.stack,
          errorType: (error as any).errorType,
        },
      ],
      status: ActionStatus.Failure,
    };
  }

  public static getEntityError(message: string): ActionResponse {
    return this.getErrorResponse(new EntityError(message));
  }

  public static getImportTableError(message: string): ActionResponse {
    return this.getErrorResponse(new ImportTableError(message));
  }

  public static getModelQueryError(message: string): ActionResponse {
    return this.getErrorResponse(new ModelQueryError(message));
  }
}
