// TODO: move this file once other parts are merged

import {
  createQueryServiceExport,
  RpcStatus,
  V1ExportFormat,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { createMutation, CreateMutationOptions } from "@tanstack/svelte-query";
import type { MutationFunction } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export const DownloadFormatMap: Record<string, V1ExportFormat> = {
  csv: V1ExportFormat.EXPORT_FORMAT_CSV,
  parquet: V1ExportFormat.EXPORT_FORMAT_PARQUET,
  xlsx: V1ExportFormat.EXPORT_FORMAT_XLSX,
};

export type DownloadReportRequest = {
  instanceId: string;
  format: string;
  bakedQuery: string;
};

export function createDownloadReportMutation<
  TError = RpcStatus,
  TContext = unknown
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<Promise<void>>,
    TError,
    { data: DownloadReportRequest },
    TContext
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  const exporter = createQueryServiceExport();

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: DownloadReportRequest }
  > = async (props) => {
    const { data } = props ?? {};
    if (!data.instanceId) throw new Error("Missing instanceId");

    const exportResp = await get(exporter).mutateAsync({
      instanceId: data.instanceId,
      data: {
        format:
          DownloadFormatMap[data.format] ??
          V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED,
        bakedQuery: data.bakedQuery,
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
