import { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
import {
  commonOptions,
  getFilterOptions,
} from "@rilldata/web-common/features/canvas/components/util";
import type { InputParams } from "@rilldata/web-common/features/canvas/inspector/types";
import type { LeaderboardState } from "@rilldata/web-common/features/dashboards/leaderboard/types";
import {
  SortDirection,
  SortType,
} from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import {
  isValueBasedSort,
  toggleSortDirection,
} from "@rilldata/web-common/features/dashboards/state-managers/actions/sorting";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable, type Writable } from "svelte/store";
import type { CanvasEntity, ComponentPath } from "../../stores/canvas-entity";
import type {
  CanvasComponentType,
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";
import Leaderboard from "./LeaderboardDisplay.svelte";

export { default as Leaderboard } from "./LeaderboardDisplay.svelte";

export const defaultComparisonOptions: ComponentComparisonOptions[] = [
  "delta",
  "percent_change",
];

export interface LeaderboardSpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measures: string[];
  dimensions: string[];
  num_rows: number;
}

export class LeaderboardComponent extends BaseCanvasComponent<LeaderboardSpec> {
  minSize = { width: 3, height: 3 };
  defaultSize = { width: 6, height: 3 };
  resetParams = ["measures", "dimensions"];
  type: CanvasComponentType = "leaderboard";
  component = Leaderboard;
  leaderboardState: Writable<LeaderboardState>;

  constructor(resource: V1Resource, parent: CanvasEntity, path: ComponentPath) {
    const defaultSpec: LeaderboardSpec = {
      metrics_view: "",
      measures: [],
      dimensions: [],
      num_rows: 7,
    };
    super(resource, parent, path, defaultSpec);

    const { measures } = get(this.specStore);
    this.leaderboardState = writable({
      sortType: SortType.VALUE,
      sortDirection: SortDirection.DESCENDING,
      leaderboardSortByMeasureName: measures?.[0] ?? null,
    });

    this.specStore.subscribe((spec) => {
      this.validateAndResetSortMeasure(spec);
    });
  }

  isValid(spec: LeaderboardSpec): boolean {
    return typeof spec.metrics_view === "string";
  }

  getExploreTransformerProperties(): Partial<ExploreState> {
    const spec = get(this.specStore);
    const leaderboardState = get(this.leaderboardState);
    return {
      visibleMeasures: spec.measures,
      visibleDimensions: spec.dimensions,
      activePage: DashboardState_ActivePage.DEFAULT,
      allMeasuresVisible: false,
      allDimensionsVisible: false,
      leaderboardSortByMeasureName:
        leaderboardState.leaderboardSortByMeasureName || spec.measures[0],
      leaderboardMeasureNames: spec.measures,
      leaderboardShowContextForAllMeasures: true,
      dashboardSortType: leaderboardState.sortType,
      sortDirection: leaderboardState.sortDirection,
    };
  }

  inputParams(): InputParams<LeaderboardSpec> {
    return {
      options: {
        metrics_view: { type: "metrics", label: "Metrics view" },
        measures: {
          type: "multi_fields",
          meta: { allowedTypes: ["measure"] },
          label: "Measures",
        },
        dimensions: {
          type: "multi_fields",
          meta: { allowedTypes: ["dimension"] },
          label: "Dimensions",
        },
        num_rows: { type: "number", label: "Number of rows" },
        ...commonOptions,
      },
      filter: getFilterOptions(),
    };
  }

  static newComponentSpec(
    metricsViewName: string,
    metricsViewSpec: V1MetricsViewSpec | undefined,
  ): LeaderboardSpec {
    const measures =
      metricsViewSpec?.measures?.slice(0, 1).map((m) => m.name as string) ?? []; // TODO: change to 3

    const dimensions =
      metricsViewSpec?.dimensions
        ?.slice(0, 3)
        .map((d) => d.name || (d.column as string)) ?? [];

    return {
      metrics_view: metricsViewName,
      measures,
      dimensions,
      num_rows: 7,
    };
  }

  validateAndResetSortMeasure = (spec: LeaderboardSpec) => {
    const state = get(this.leaderboardState);
    const { measures } = spec;
    if (
      measures?.length &&
      state.leaderboardSortByMeasureName &&
      !measures.includes(state.leaderboardSortByMeasureName)
    ) {
      this.leaderboardState.set({
        ...state,
        leaderboardSortByMeasureName: measures[0],
        sortType: SortType.VALUE,
        sortDirection: SortDirection.DESCENDING,
      });
    }
  };

  // Rewrite of @toggleSort from actions/sorting.ts
  toggleSort = (sortType: SortType, measureName?: string) => {
    const state = get(this.leaderboardState);

    // Handle measure name change if provided
    if (
      measureName !== undefined &&
      measureName !== state.leaderboardSortByMeasureName &&
      isValueBasedSort(sortType)
    ) {
      this.leaderboardState.set({
        ...state,
        leaderboardSortByMeasureName: measureName,
        sortType,
        sortDirection: SortDirection.DESCENDING,
      });
      return;
    }

    // Handle sort type and direction changes
    if (sortType === undefined || state.sortType === sortType) {
      this.leaderboardState.set({
        ...state,
        sortDirection: toggleSortDirection(state.sortDirection),
      });
    } else {
      this.leaderboardState.set({
        ...state,
        sortType,
        sortDirection: SortDirection.DESCENDING,
      });
    }
  };
}
