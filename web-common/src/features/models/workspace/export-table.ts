import {
  createQueryServiceExport,
  RpcStatus,
  V1ExportFormat,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { createMutation, CreateMutationOptions } from "@tanstack/svelte-query";
import type { MutationFunction } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export type ExportTableRequest = {
  instanceId: string;
  format: V1ExportFormat;
  tableName: string;
};

export function createExportTableMutation<
  TError = { response: { data: RpcStatus } },
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<Promise<void>>,
    TError,
    { data: ExportTableRequest },
    TContext
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  const exporter = createQueryServiceExport();

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: ExportTableRequest }
  > = async (props) => {
    const { data } = props ?? {};
    if (!data.instanceId) throw new Error("Missing instanceId");

    const exportResp = await get(exporter).mutateAsync({
      instanceId: data.instanceId,
      data: {
        format: data.format,
        query: {
          tableRowsRequest: {
            instanceId: data.instanceId,
            tableName: data.tableName,
          },
        },
      },
    });
    const downloadUrl = `${get(runtime).host}${exportResp.downloadUrlPath}`;
    window.open(downloadUrl, "_self");
  };

  return createMutation<
    Awaited<Promise<void>>,
    TError,
    { data: ExportTableRequest },
    TContext
  >(mutationFn, mutationOptions);
}
