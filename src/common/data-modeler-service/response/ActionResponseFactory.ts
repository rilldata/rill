import { ActionResponseMessageType } from "$common/data-modeler-service/response/ActionResponseMessage";
import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import { ImportTableError } from "$common/errors/ImportTableError";
import { ModelQueryError } from "$common/errors/ModelQueryError";
import { EntityError } from "$common/errors/EntityError";
import { ExistingEntityError } from "$common/errors/ExistingEntityError";
import type { TypedError } from "$common/errors/TypedError";

export class ActionResponseFactory {
  public static getSuccessResponse(
    message?: string,
    data?: unknown
  ): ActionResponse {
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
      ...(data !== undefined ? { data } : {}),
    };
  }
  public static getRawResponse(data?: Record<any, any>): ActionResponse {
    return {
      messages: [],
      status: ActionStatus.Success,
      ...data,
    };
  }

  public static getErrorResponse(error: Error & TypedError): ActionResponse {
    return {
      messages: [
        {
          type: ActionResponseMessageType.Error,
          message: error.message,
          stack: error.stack,
          errorType: error.errorType,
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

  public static getExisingEntityError(message: string): ActionResponse {
    return this.getErrorResponse(new ExistingEntityError(message));
  }
}
