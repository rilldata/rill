// TODO: move this file once other parts are merged

import { createQueryServiceExportReportMutation } from "@rilldata/web-common/runtime-client/v2/gen";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createMutation,
  type CreateMutationOptions,
} from "@tanstack/svelte-query";
import type { MutationFunction } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export type DownloadReportRequest = {
  reportId: string;
  executionTime: string;
  originBaseUrl: string;
  host: string;
};

export function createDownloadReportMutation<
  TError = { response: { data: RpcStatus } },
  TContext = unknown,
>(
  client: RuntimeClient,
  options?: {
    mutation?: CreateMutationOptions<
      Awaited<Promise<void>>,
      TError,
      { data: DownloadReportRequest },
      TContext
    >;
  },
) {
  const { mutation: mutationOptions } = options ?? {};
  const exporter = createQueryServiceExportReportMutation(client);

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: DownloadReportRequest }
  > = async (props) => {
    const { data } = props ?? {};

    // executionTime is an ISO string; the v2/gen layer uses fromJson internally,
    // which handles string→Timestamp conversion at runtime
    const exportResp = await get(exporter).mutateAsync({
      report: data.reportId,
      executionTime: data.executionTime as any, // eslint-disable-line @typescript-eslint/no-explicit-any
      originBaseUrl: data.originBaseUrl,
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
