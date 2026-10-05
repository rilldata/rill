import {
  createQueryServiceMetricsViewEvaluate,
  type MetricsViewSpecMeasure,
  type V1MetricsViewAggregationRequest,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import EvaluationAnswerCell from "@rilldata/web-common/features/evaluations/EvaluationAnswerCell.svelte";
import EvaluationNOULAnswerCell from "@rilldata/web-common/features/evaluations/EvaluationNOULAnswerCell.svelte";
import EvaluationChoiceAnswerCell from "@rilldata/web-common/features/evaluations/EvaluationChoiceAnswerCell.svelte";
import EvaluationScoreAnswerCell from "@rilldata/web-common/features/evaluations/EvaluationScoreAnswerCell.svelte";

export function getEvaluationMeasuresQuery(
  runtimeClient: RuntimeClient,
  aggregationRequest: V1MetricsViewAggregationRequest,
  evalMeasureNames: string[],
) {
  return createQueryServiceMetricsViewEvaluate(
    runtimeClient,
    {
      metricsViewName: aggregationRequest.metricsView,
      query: aggregationRequest,
      measures: evalMeasureNames,
    },
    {
      query: {
        enabled: evalMeasureNames.length > 0,
      },
    },
  );
}

const Components = {
  noul: EvaluationNOULAnswerCell,
  choice: EvaluationChoiceAnswerCell,
  score: EvaluationScoreAnswerCell,
};
export function getEvaluationMeasureColumns(
  evalMeasures: MetricsViewSpecMeasure[],
) {
  return evalMeasures.map((m) => {
    const cellComponent =
      Components[Object.keys(m.evalQuestion ?? {})[0]] ?? EvaluationAnswerCell;

    return {
      name: m.name!,
      type: "VARCHAR",
      label: getMeasureDisplayName(m),
      enableResize: true,
      enableSorting: false,
      cellComponent,
    };
  });
}
