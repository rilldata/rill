import {
  RpcStatus,
  V1PutFileAndReconcileRequest,
  V1PutFileAndReconcileResponse,
  runtimeServicePutFileAndReconcile,
} from "@rilldata/web-common/runtime-client";
import {
  CreateMutationOptions,
  MutationFunction,
  createMutation,
} from "@tanstack/svelte-query";

const RILL_COMPILER = "rill-beta";
const RILL_PROJECT_TITLE = "Untitled Rill Project";

export const useInitializeProjectFile = <
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<Promise<V1PutFileAndReconcileResponse>>,
    TError,
    { data: V1PutFileAndReconcileRequest },
    TContext
  >;
}) => {
  const { mutation: mutationOptions } = options ?? {};

  const mutationFn: MutationFunction<
    Awaited<Promise<V1PutFileAndReconcileResponse>>,
    { data: V1PutFileAndReconcileRequest }
  > = async (props) => {
    const { data } = props ?? {};

    const res = await runtimeServicePutFileAndReconcile({
      instanceId: data.instanceId,
      path: "rill.yaml",
      blob: `title: ${RILL_PROJECT_TITLE}\ncompiler: ${RILL_COMPILER}`,
    });

    return res;
  };

  return createMutation<
    Awaited<Promise<V1PutFileAndReconcileResponse>>,
    TError,
    { data: V1PutFileAndReconcileRequest },
    TContext
  >(mutationFn, mutationOptions);
};
