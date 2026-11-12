import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params";
import {
  V1Operation,
  type V1Expression,
} from "@rilldata/web-common/runtime-client";
import { TIME_FILTER_INHERIT } from "../components/time-filters";
import { ControlKind, type AppliedControl } from "./controls";

/**
 * The filter that a component's filter controls add to its queries, or undefined if none is set.
 * A control filters the dimension named by its target, or by its `fields` entry for the
 * component's metrics view. hasDimension tells whether that metrics view has a dimension;
 * a control whose dimension it lacks doesn't apply.
 */
export function controlFiltersExpression(
  controls: AppliedControl[],
  metricsViewName: string,
  hasDimension: (name: string) => boolean,
): V1Expression | undefined {
  const exprs: V1Expression[] = [];
  for (const { control, values } of controls) {
    if (control.kind !== ControlKind.Filter || !values?.length) continue;
    const dimension = control.fields?.[metricsViewName] ?? control.target;
    if (!dimension || !hasDimension(dimension)) continue;
    exprs.push(createInExpression(dimension, values));
  }
  return exprs.length ? createAndExpression(exprs) : undefined;
}

/**
 * `where` narrowed by a control filter. Every predicate of both is kept: unlike mergeFilters,
 * which rebuilds the expression from the kinds of predicate it understands, this only adds
 * the control's predicates to the conjunction.
 */
export function andControlFilters(
  where: V1Expression | undefined,
  controlWhere: V1Expression,
): V1Expression {
  if (!where) return controlWhere;
  const controlExprs = controlWhere.cond?.exprs ?? [];
  if (where.cond?.op === V1Operation.OPERATION_AND) {
    return createAndExpression([...(where.cond.exprs ?? []), ...controlExprs]);
  }
  return createAndExpression([where, ...controlExprs]);
}

/**
 * A component's `time_filters` with its time controls applied: a time range control replaces
 * `tr`, and a time grain control replaces `grain`. A range picked through a control keeps the
 * canvas comparison unless the YAML says otherwise, like a range picked in the editor does.
 */
export function applyTimeControls(
  timeFilters: string,
  controls: AppliedControl[],
): URLSearchParams {
  const params = new URLSearchParams(timeFilters);
  for (const { control, values } of controls) {
    const value = values?.[0];
    if (!value) continue;
    if (control.kind === ControlKind.TimeRange) {
      params.set(ExploreStateURLParams.TimeRange, value);
      if (!params.has(ExploreStateURLParams.ComparisonTimeRange)) {
        params.set(
          ExploreStateURLParams.ComparisonTimeRange,
          TIME_FILTER_INHERIT,
        );
      }
    } else if (control.kind === ControlKind.TimeGrain) {
      params.set(ExploreStateURLParams.TimeGrain, value);
    }
  }
  return params;
}
