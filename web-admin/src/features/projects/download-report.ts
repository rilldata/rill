// TODO: move this file once other parts are merged

import {
  createQueryServiceExportReport,
  type RpcStatus,
} from "@rilldata/web-common/runtime-client";
import {
  createMutation,
  type CreateMutationOptions,
} from "@tanstack/svelte-query";
import type { MutationFunction } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export type DownloadReportRequest = {
  instanceId: string;
  reportId: string;
  executionTime: string;
  originBaseUrl: string;
  host: string;
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
        executionTime: data.executionTime,
        originBaseUrl: data.originBaseUrl,
      },
    });
    const downloadUrl = `${data.host}${exportResp.downloadUrlPath}`;
    window.open(downloadUrl, "_self");
  };

  return createMutation<
    Awaited<Promise<void>>,
    TError,
    { data: DownloadReportRequest },
    TContext
  >({ mutationFn, ...mutationOptions });
}
