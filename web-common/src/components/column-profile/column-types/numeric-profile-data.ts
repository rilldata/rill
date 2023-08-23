import type { ColumnsProfileDataUpdate } from "@rilldata/web-common/components/column-profile/columns-profile-data";
import { chooseBetweenDiagnosticAndStatistical } from "@rilldata/web-common/components/column-profile/utils";
import {
  NumericHistogramBinsBin,
  QueryServiceColumnNumericHistogramHistogramMethod,
  V1HistogramMethod,
} from "@rilldata/web-common/runtime-client";
import type { BatchedRequest } from "@rilldata/web-common/runtime-client/batched-request";
import { getPriority } from "@rilldata/web-common/runtime-client/http-request-queue/priorities";

export async function loadColumnHistogram(
  instanceId: string,
  tableName: string,
  columnName: string,
  isFloat: boolean,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  let histogramData: Array<NumericHistogramBinsBin>;

  const diagnosticHistogramPromise = getHistogramData(
    instanceId,
    tableName,
    columnName,
    QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_DIAGNOSTIC,
    batchedRequest
  );
  if (isFloat) {
    const fdHistogramPromise = getHistogramData(
      instanceId,
      tableName,
      columnName,
      QueryServiceColumnNumericHistogramHistogramMethod.HISTOGRAM_METHOD_FD,
      batchedRequest
    );

    const [diagnosticHistogram, fdHistogram] = await Promise.all([
      diagnosticHistogramPromise,
      fdHistogramPromise,
    ]);
    histogramData = chooseBetweenDiagnosticAndStatistical(
      diagnosticHistogram,
      fdHistogram
    );
  } else {
    histogramData = await diagnosticHistogramPromise;
  }

  update((state) => {
    if (!state.profiles[columnName]) return;
    state.profiles[columnName].histogram = histogramData;
    return state;
  });
}

export async function loadDescriptiveStatistics(
  instanceId: string,
  tableName: string,
  columnName: string,
  batchedRequest: BatchedRequest,
  update: ColumnsProfileDataUpdate
) {
  const descriptiveStatistics = await batchedRequest.addReq(
    {
      columnDescriptiveStatisticsRequest: {
        instanceId,
        tableName,
        columnName,
        priority: getPriority("descriptive-statistics"),
      },
    },
    (data) =>
      data.columnDescriptiveStatisticsResponse.numericSummary.numericStatistics
  );

  update((state) => {
    if (!state.profiles[columnName]) return;
    state.profiles[columnName].descriptiveStatistics = descriptiveStatistics;
    return state;
  });
}

function getHistogramData(
  instanceId: string,
  tableName: string,
  columnName: string,
  histogramMethod: V1HistogramMethod,
  batchedRequest: BatchedRequest
) {
  return batchedRequest.addReq(
    {
      columnNumericHistogramRequest: {
        instanceId,
        tableName,
        columnName,
        histogramMethod,
        priority: getPriority("numeric-histogram"),
      },
    },
    (data) =>
      data.columnNumericHistogramResponse?.numericSummary?.numericHistogramBins
        ?.bins
  );
}
