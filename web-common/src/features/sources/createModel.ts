import { createModel } from "@rilldata/web-common/features/models/createModel";
import { getName } from "@rilldata/web-local/lib/util/incrementName";
import type { QueryClient, UseMutationResult } from "@sveltestack/svelte-query";
import { notifications } from "../../components/notifications";
import type { V1PutFileAndReconcileResponse } from "../../runtime-client";

export async function createModelFromSource(
  queryClient: QueryClient,
  instanceId: string,
  modelNames: Array<string>,
  sourceName: string,
  sourceNameInQuery: string,
  createModelMutation: UseMutationResult<V1PutFileAndReconcileResponse>, // TODO: type
  setAsActive = true
): Promise<string> {
  const newModelName = getName(`${sourceName}_model`, modelNames);
  await createModel(
    queryClient,
    instanceId,
    newModelName,
    createModelMutation,
    `select * from ${sourceNameInQuery}`,
    setAsActive
  );
  notifications.send({
    message: `Queried ${sourceNameInQuery} in workspace`,
  });
  return newModelName;
}
