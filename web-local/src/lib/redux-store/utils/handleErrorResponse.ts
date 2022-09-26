import type { ActionResponse } from "$web-local/common/data-modeler-service/response/ActionResponse";
import { ActionStatus } from "$web-local/common/data-modeler-service/response/ActionResponse";

/** this is currently a no-op */
export function handleErrorResponse(actionResponse: ActionResponse) {
  if (!actionResponse || actionResponse.status === ActionStatus.Success) return;
}
