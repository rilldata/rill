import type { PivotState } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import { ExploreMetricsViewMetadata } from "@rilldata/web-common/features/dashboards/stores/ExploreMetricsViewMetadata.ts";
import {
  Filters,
  type FiltersState,
  getEmptyFilterState,
} from "@rilldata/web-common/features/dashboards/stores/Filters.ts";
import {
  getEmptyTimeControlState,
  TimeControls,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { PivotStore } from "@rilldata/web-common/features/scheduled-reports/pivot-dashboard/pivot-store.ts";

type ReportPivotRenderData = {
  metadata: ExploreMetricsViewMetadata;
  filters: Filters;
  timeControls: TimeControls;
  pivotStore: PivotStore;
};
export class ReportPivotRenderStore {
  private dataStore = new Map<string, ReportPivotRenderData>();

  public constructor(
    private readonly initFilters: FiltersState = getEmptyFilterState(),
    private readonly initTimeControls: TimeControlState = getEmptyTimeControlState(),
    private readonly initPivotState: PivotState = getEmptyPivotState(),
  ) {}

  public get(
    instanceId: string,
    metricsViewName: string,
    exploreName: string,
    canvasName: string,
    useInit: boolean,
  ): ReportPivotRenderData {
    const key = `${instanceId}__${metricsViewName}__${exploreName || canvasName}`;
    if (this.dataStore.has(key)) {
      return this.dataStore.get(key)!;
    }

    const metadata = new ExploreMetricsViewMetadata(
      instanceId,
      metricsViewName,
      exploreName,
    );

    const initFilters = useInit ? this.initFilters : getEmptyFilterState();
    const filters = new Filters(metadata, initFilters);

    const initTimeControls = useInit
      ? this.initTimeControls
      : getEmptyTimeControlState();
    const timeControls = new TimeControls(metadata, initTimeControls);

    const initPivotState = useInit ? this.initPivotState : getEmptyPivotState();
    const pivotStore = new PivotStore(initPivotState);

    const renderData: ReportPivotRenderData = {
      metadata,
      filters,
      timeControls,
      pivotStore,
    };
    this.dataStore.set(key, renderData);
    return renderData;
  }
}

function getEmptyPivotState(): PivotState {
  return {
    rows: [],
    columns: [],
    sorting: [],
    expanded: {},
    columnPage: 1,
    rowPage: 1,
    enableComparison: true,
    activeCell: null,
    tableMode: "flat",
  };
}
