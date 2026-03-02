// TODO: move this file once other parts are merged

import { Timestamp } from "@bufbuild/protobuf";
import {
  queryServiceExportReport,
  type RpcStatus,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createMutation,
  type CreateMutationOptions,
} from "@tanstack/svelte-query";
import type { MutationFunction } from "@tanstack/svelte-query";

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

  const mutationFn: MutationFunction<
    Awaited<Promise<void>>,
    { data: DownloadReportRequest }
  > = async (props) => {
    const { data } = props ?? {};

    const exportResp = await queryServiceExportReport(client, {
      report: data.reportId,
      executionTime: Timestamp.fromJson(data.executionTime),
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
