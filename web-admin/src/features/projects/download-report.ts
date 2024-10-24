// TODO: move this file once other parts are merged

import {
  createQueryServiceExportReport,
  type RpcStatus,
  V1ExportFormat,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  createMutation,
  type CreateMutationOptions,
} from "@tanstack/svelte-query";
import type { MutationFunction } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export type DownloadReportRequest = {
  instanceId: string;
  reportId: string;
  format: V1ExportFormat;
  executionTime: string;
  limit?: string;
};

export function createDownloadReportMutation<
  TError = { response: { data: RpcStatus } },
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<Promise<void>>,
    TError,
    { data: DownloadReportRequest },
    TContext
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  const exporter = createQueryServiceExportReport();

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: DownloadReportRequest }
  > = async (props) => {
    const { data } = props ?? {};
    if (!data.instanceId) throw new Error("Missing instanceId");

    const exportResp = await get(exporter).mutateAsync({
      instanceId: data.instanceId,
      report: data.reportId,
      data: {
        format: data.format,
        executionTime: data.executionTime,
        limit: data.limit,
      },
    });
    const downloadUrl = `${get(runtime).host}${exportResp.downloadUrlPath}`;
    window.open(downloadUrl, "_self");
  };

  return createMutation<
    Awaited<Promise<void>>,
    TError,
    { data: DownloadReportRequest },
    TContext
  >(mutationFn, mutationOptions);
}
