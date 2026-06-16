import { goto } from "$app/navigation";
import { page } from "$app/stores";
import {
  useCanvas,
  type CanvasResponse,
} from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  V1ExploreComparisonMode,
  type V1CanvasPreset,
  type V1CanvasSpec,
  type V1ComponentSpecRendererProperties,
  type V1MetricsView,
  type V1MetricsViewSpec,
  type V1Resource,
  type V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import {
  derived,
  get,
  writable,
  type Readable,
  type Unsubscriber,
} from "svelte/store";
import { parseDocument, YAMLMap, isMap, type Pair } from "yaml";
import type { FileArtifact } from "../../entity-management/file-artifact";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { MetricsViewSelectors } from "../../metrics-views/metrics-view-selectors";
import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
import type { CanvasComponentType, ComponentSpec } from "../components/types";
import {
  COMPONENT_CLASS_MAP,
  createComponent,
  isChartComponentType,
  isTableComponentType,
} from "../components/util";
import { FilterManager, flattenExpression } from "./filter-manager";
import { getFilterParam } from "./filter-state";
import { Grid } from "./grid";
import { getComparisonTypeFromRangeString } from "./time-state";
import { TimeManager } from "./time-manager";
import { Theme } from "../../themes/theme";
import { createResolvedThemeStore } from "../../themes/selectors";
import { ExploreStateURLParams } from "../../dashboards/url-state/url-params";
import { DEFAULT_DASHBOARD_WIDTH } from "../layout-util";
import { createCustomMapStore } from "@rilldata/web-common/lib/custom-map-store";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryServiceConvertExpressionToMetricsSQL } from "@rilldata/web-common/runtime-client";

export const lastVisitedState = new Map<string, string>();

// Store for managing URL search parameters
// Which may be in the URL or in the Canvas YAML
// Set returns a boolean indicating whether the value was set
export type SearchParamsStore = {
  subscribe: (run: (value: URLSearchParams) => void) => Unsubscriber;
  set: (
    map: Map<string, string | undefined>,
    checkIfSet?: boolean,
    replaceState?: boolean,
    prefixes?: string[],
  ) => boolean;
  clearAll: () => void;
};

export class CanvasEntity {
  componentsStore = createCustomMapStore<BaseCanvasComponent>();
  _rows: Grid = new Grid(this);

  // Time state controls
  timeManager: TimeManager;

  // Dimension and measure filter state
  filterManager: FilterManager;

  // Metrics view selectors
  metricsView: MetricsViewSelectors;

  fileArtifact: FileArtifact | undefined;

  selectedComponent = writable<string | null>(null);
  activeComponent = writable<string | null>(null);
  parsedContent: Readable<ReturnType<typeof parseDocument>>;
  public specStore: CanvasSpecResponseStore;
  // Tracks whether the canvas been loaded (and rows processed) for the first time
  firstLoad = writable(true);
  themeName = writable<string | undefined>(undefined);
  theme: Readable<Theme | undefined>;
  unsubscriber: Unsubscriber;
  private searchParams = writable<URLSearchParams>(new URLSearchParams());
  // This may sometimes be false due to discrepancy between two different ways
  // of storing the same state in the URL namely dimension IN (['value']) vs  dimension IN ('value')
  defaultUrlParamsStore = writable<URLSearchParams>(new URLSearchParams());
  viewingDefaultsStore: Readable<boolean>;
  filtersEnabledStore = writable<boolean>(true);
  _embeddedTheme = writable<V1ThemeSpec | undefined>(undefined);
  _metricsViews = writable<Record<string, V1MetricsView | undefined>>({});
  bannerStore = writable<string | undefined>(undefined);
  _maxWidth = writable<number>(DEFAULT_DASHBOARD_WIDTH);
  titleStore = writable<string>("");

  // This is to skip processing the spec the first time the store updates with a value
  // We've already called it as part of the constructor
  firstTimeLoad = true;

  constructor(
    public name: string,
    public instanceId: string,
    private spec: CanvasResponse,
    readonly client: RuntimeClient,
    public allowUnvalidatedSpec = false,
  ) {
    this.specStore = useCanvas(
      client,
      name,
      {},
      queryClient,
      allowUnvalidatedSpec,
    );

    // This will be deprecated soon - bgh
    const searchParamsStore: SearchParamsStore = (() => {
      return {
        subscribe: this.searchParams.subscribe,
        set: (
          map: Map<string, string>,
          checkIfSet = false,
          replaceState = false,
        ) => {
          const existingParams = new URLSearchParams(window.location.search);

          map.forEach((value, key) => {
            if (checkIfSet && existingParams.has(key)) return false;

            if (value === undefined || value === null || value === "") {
              existingParams.delete(key);
            } else {
              existingParams.set(key, value);
            }
          });

          goto(`?${existingParams.toString()}`, { replaceState }).catch(
            console.error,
          );
          return true;
        },
        clearAll: () => {
          const url = get(page).url;
          url.searchParams.forEach((_, effectiveKey) => {
            url.searchParams.delete(effectiveKey);
          });

          goto(url.toString(), { replaceState: true }).catch(console.error);
        },
      };
    })();

    this.theme = createResolvedThemeStore(
      this.themeName,
      this.specStore,
      this.client,
    );

    this.timeManager = new TimeManager(searchParamsStore, this);

    // Let the embed layer (CanvasDashboardEmbed) drive themeName;
    // initialise with no override here so createResolvedThemeStore falls
    // back to the dashboard's own theme from the spec unless an embed
    // override is applied.
    this.themeName.set(undefined);

    this.processSpec(this.spec);

    this.metricsView = new MetricsViewSelectors(
      this.client,
      this._metricsViews,
    );

    this.unsubscriber = this.specStore.subscribe(({ data }) => {
      if (this.firstTimeLoad) {
        this.firstTimeLoad = false;
        return;
      }
      if (data) {
        this.processSpec(data);
      }
    });

    this.viewingDefaultsStore = derived(
      [
        this.searchParams,
        this.defaultUrlParamsStore,
        this.filterManager.pinnedFilterKeysStore,
        this.filterManager.defaultPinnedFilterKeysStore,
        this.filterManager.requiredFilterKeysStore,
        this.filterManager.defaultRequiredFilterKeysStore,
      ],
      ([
        $searchParams,
        $defaultUrlParams,
        pinnedFilters,
        defaultPinnedFilterKeys,
        requiredFilters,
        defaultRequiredFilterKeys,
      ]) => {
        if (
          defaultPinnedFilterKeys.symmetricDifference(pinnedFilters).size > 0
        ) {
          return false;
        }
        if (
          defaultRequiredFilterKeys.symmetricDifference(requiredFilters).size >
          0
        ) {
          return false;
        }
        if ($defaultUrlParams.size === 0) {
          return false;
        }

        for (const [key, value] of $defaultUrlParams.entries()) {
          if ($searchParams.get(key) !== value) {
            // Ignore time range if not set
            if (
              $searchParams.get(key) === null &&
              key === ExploreStateURLParams.TimeRange
            ) {
              continue;
            }
            return false;
          }
        }
        for (const [key, value] of $searchParams.entries()) {
          if ($defaultUrlParams.get(key) !== value) {
            return false;
          }
        }
        return true;
      },
    );
  }

  checkAndSetMaxWidth = ({ maxWidth }: V1CanvasSpec) => {
    const currentValue = get(this._maxWidth);

    if (maxWidth && maxWidth !== currentValue) {
      this._maxWidth.set(maxWidth);
    }
  };

  checkAndSetHasBanner = ({ banner }: V1CanvasSpec) => {
    const currentValue = get(this.bannerStore);
    if (banner !== currentValue) {
      this.bannerStore.set(banner);
    }
  };

  checkAndSetDefaultParams = (defaultPreset: V1CanvasPreset) => {
    const defaultSearchParams = getDefaults(defaultPreset);
    const currentDefaultParams = get(this.defaultUrlParamsStore);
    if (defaultSearchParams.toString() !== currentDefaultParams.toString()) {
      this.defaultUrlParamsStore.set(defaultSearchParams);
    }
  };

  checkAndSetFilterEnabled = ({ filtersEnabled }: V1CanvasSpec) => {
    if (filtersEnabled === undefined) {
      filtersEnabled = true;
    }

    const currentValue = get(this.filtersEnabledStore);
    if (filtersEnabled !== currentValue) {
      this.filtersEnabledStore.set(filtersEnabled);
    }
  };

  checkAndSetEmbeddedTheme = ({ embeddedTheme }: V1CanvasSpec) => {
    const currentValue = get(this._embeddedTheme);
    if (embeddedTheme !== currentValue) {
      this._embeddedTheme.set(embeddedTheme);
    }
  };

  checkAndSetFileArtifact = (filePath: string | undefined) => {
    if (!filePath) return;
    if (!this.fileArtifact) {
      const fileArtifact = fileArtifacts.getFileArtifact(filePath);

      if (!fileArtifact) {
        return;
      }

      this.fileArtifact = fileArtifact;

      if (!this.parsedContent) {
        this.parsedContent = derived(
          fileArtifact.editorContent,
          (editorContent) => {
            const parsed = parseDocument(editorContent ?? "");
            return parsed;
          },
        );
      }
    }
  };

  processSpec = (response: CanvasResponse) => {
    const { canvas, components, metricsViews, filePath } = response;
    const validSpec = canvas;
    if (!validSpec) return;

    if (metricsViews) this._metricsViews.set(metricsViews);

    this.checkAndSetFilterEnabled(validSpec);
    this.checkAndSetFileArtifact(filePath);
    this.checkAndSetDefaultParams(validSpec.defaultPreset ?? {});
    this.checkAndSetEmbeddedTheme(validSpec);
    this.checkAndSetHasBanner(validSpec);
    this.checkAndSetMaxWidth(validSpec);

    this.timeManager.onSpecChange(response);

    this.titleStore.set(validSpec.displayName ?? "");

    const defaultPreset = validSpec?.defaultPreset ?? {};
    const filterExpressions = defaultPreset.filterExpr ?? {};
    const pinnedFilters = validSpec?.pinnedFilters ?? [];
    const requiredFilters = validSpec?.requiredFilters ?? [];

    if (metricsViews) {
      if (this.filterManager) {
        this.filterManager.updateConfig(
          metricsViews,
          pinnedFilters,
          filterExpressions,
          requiredFilters,
        );
      } else {
        this.filterManager = new FilterManager(
          metricsViews,
          this.instanceId,
          pinnedFilters,
          filterExpressions,
          requiredFilters,
        );
        // Clears the active component when a global filter changes through
        // FilterManager.actions.* (user-driven filter UI). Pivot click-to-filter
        // bypasses actions and mutates FilterState directly, so it does NOT
        // trigger this callback; see pivot-click-to-filter.ts for details.
        this.filterManager.onFilterChange = () => this.clearActiveComponent();
      }
    } else {
      // need to find a better way to initialize this in certain contextx - bgh
      this.filterManager = new FilterManager({}, "", [], {});
      this.filterManager.onFilterChange = () => this.clearActiveComponent();
    }

    this.processRows({ canvas, components, metricsViews, filePath });
  };

  saveDefaultFilters = async () => {
    // Temporary solution to wait for any pending filter updates to propagate
    // This happens when a user has changed a filter but not yet
    // clicked out of the filter input box to save the values
    await new Promise((resolve) => {
      setTimeout(resolve, 100);
    });

    const pinnedFilters = get(this.filterManager.pinnedFilterKeysStore);
    const requiredFilters = get(this.filterManager.requiredFilterKeysStore);

    // Persist pinned and required independently. Render-time treats a filter as
    // visible whenever it's in either set, so we don't dedupe here: doing so
    // would silently drop the pin flag if a user later toggled required off.
    const pinnedNames = Array.from(pinnedFilters).map((f) => f.split("::")[1]);
    const requiredNames = Array.from(requiredFilters).map(
      (f) => f.split("::")[1],
    );
    const timeRange = get(this.timeManager.state.rangeStore);
    const comparisonOn = get(this.timeManager.state.showTimeComparisonStore);

    const metricsViewFilters = get(this.filterManager.metricsViewFilters);
    const filterNames = Array.from(metricsViewFilters.keys());
    const promises = Array.from(metricsViewFilters.values()).map((filters) => {
      const parsed = get(filters.parsed);
      return queryClient.fetchQuery({
        queryKey: [
          "resolve-metrics-view-filter-expression",
          this.instanceId,
          parsed.where,
        ],
        queryFn: () =>
          queryServiceConvertExpressionToMetricsSQL(this.client, {
            expression: parsed.where as any,
          }),
      });
    });

    const responses = await Promise.all(promises);

    const filterMap = new YAMLMap();
    responses.forEach((response, index) => {
      if (!response.sql) return;
      filterMap.add({
        key: filterNames[index],
        value: response.sql,
      });
    });

    // Read the YAML document AFTER the async work so any component edits
    // that landed in editorContent during the await are preserved. Reading
    // before the await captures a stale Document whose round-trip back to
    // disk would drop those components. The mutate-and-write below runs
    // synchronously to keep the read-modify-write atomic.
    const yaml = get(this.parsedContent);

    setOrDeleteFilterList(yaml, "pinned", pinnedNames);
    setOrDeleteFilterList(yaml, "required", requiredNames);

    if (
      yaml.get("filters") instanceof YAMLMap &&
      (yaml.get("filters") as YAMLMap).items.length === 0
    ) {
      try {
        yaml.deleteIn(["filters"]);
      } catch {
        // no-op
      }
    }

    if (timeRange) {
      yaml.setIn(["defaults", "time_range"], timeRange);
    }

    if (comparisonOn) {
      yaml.setIn(["defaults", "comparison_mode"], "time");
    } else {
      try {
        yaml.deleteIn(["defaults", "comparison_mode"]);
      } catch {
        // no-op
      }
    }

    yaml.setIn(["defaults", "filters"], filterMap);

    if (yaml.contents && isMap(yaml.contents)) {
      yaml.contents.items.sort(customKeySort);
    }

    const newContent = yaml.toString();

    this.fileArtifact?.updateEditorContent(newContent, false, true);

    // Navigate to new URL after update
    let firstUpdate = true;
    const unsub = this.defaultUrlParamsStore.subscribe((newParam) => {
      if (firstUpdate) {
        firstUpdate = false;
        return;
      }
      goto(`?${newParam.toString()}`).catch(console.error);
      unsub();
    });
  };

  clearDefaultFilters = () => {
    const yaml = get(this.parsedContent);

    try {
      yaml.deleteIn(["defaults", "filters"]);
      yaml.deleteIn(["defaults", "time_range"]);
      yaml.deleteIn(["defaults", "comparison_mode"]);
    } catch {
      // no-op
    }

    const defaultSize = yaml.get("defaults").items.length;
    if (defaultSize === 0) {
      try {
        yaml.delete("defaults");
      } catch {
        // no-op
      }
    }

    try {
      yaml.delete("filters");
    } catch {
      // no-op
    }

    const newContent = yaml.toString();

    this.fileArtifact?.updateEditorContent(newContent, false, true);
  };

  onUrlChange = async ({
    url: { searchParams, pathname },
    projectId,
  }: {
    url: URL;
    projectId?: string;
  }) => {
    const redirected = await this.handleCanvasRedirect({
      canvasName: this.name,
      searchParams,
      pathname,

      projectId,
    });

    if (redirected) return;

    this.filterManager.onUrlChange(searchParams);
    this.searchParams.set(searchParams);
    this.saveSnapshot(searchParams.toString());
    this.timeManager.state.onUrlChange(searchParams);
  };

  // Not currently being used
  unsubscribe = () => {
    // this.unsubscriber();
  };

  handleCanvasRedirect = async ({
    canvasName,
    searchParams,
    pathname,
    projectId,
  }: {
    canvasName: string;
    searchParams: URLSearchParams;
    pathname: string;

    projectId?: string;
  }) => {
    // If there are no URL params, check for last visited state or home bookmark
    if (searchParams.size === 0) {
      const snapshotSearchParams = lastVisitedState.get(canvasName);

      // First priority
      if (snapshotSearchParams) {
        await goto(`?${snapshotSearchParams}`, { replaceState: true });
        return true;
      }

      // Second priority
      const deployed = projectId;

      if (deployed) {
        let homeBookmarkUrlSearch: string | undefined = undefined;
        try {
          // Only gets imported in admin context
          const { getAdminServiceListBookmarksQueryOptions } = await import(
            "@rilldata/web-admin/client"
          );

          const queryOptions = getAdminServiceListBookmarksQueryOptions({
            projectId,
            resourceKind: ResourceKind.Canvas,
            resourceName: canvasName,
          });

          const response = await queryClient.fetchQuery(queryOptions);

          const homeBookmark = response.bookmarks?.find(
            (bookmark) => bookmark.default,
          );

          homeBookmarkUrlSearch = homeBookmark?.urlSearch;
        } catch (e) {
          console.error("Error fetching bookmarks for canvas redirect:", e);
        }

        if (homeBookmarkUrlSearch) {
          await goto(homeBookmarkUrlSearch, { replaceState: true });
          return true;
        }
      }

      // Third priority
      const defaultParamsString = get(this.defaultUrlParamsStore).toString();

      if (defaultParamsString) {
        await goto(`?${defaultParamsString}`, {
          replaceState: true,
        });
        return true;
      }
    } else if (searchParams.get("default")) {
      // If the default parameter exists, we clear last visited state and redirect to clean URL
      lastVisitedState.set(canvasName, "");

      await goto(pathname, { replaceState: true });
      return true;
    }
  };

  saveSnapshot = (filterState: string) => {
    lastVisitedState.set(this.name, filterState);
  };

  duplicateItem = (id: string) => {
    const component = this.componentsStore.getNonReactive(id);
    if (!component) return;
    const { pathInYAML, type, resource } = component;
    const [, rowIndex, , columnIndex] = pathInYAML;
    const path = constructPath(rowIndex, columnIndex, type);

    const existingResource = get(resource);

    const metricsViewName = existingResource?.component?.state?.validSpec
      ?.rendererProperties?.metrics_view as string | undefined;

    if (!metricsViewName) {
      throw new Error("No metrics view name found");
    }

    const metricsViewSpec = get(
      this.metricsView.getMetricsViewFromName(metricsViewName),
    ).metricsView;

    if (!metricsViewSpec) {
      throw new Error("No metrics view spec found");
    }

    const newResource = this.createOptimisticResource({
      type,
      row: rowIndex + 1,
      column: columnIndex,
      metricsViewName,
      metricsViewSpec,
    });

    const newComponent = createComponent(newResource, this, path);
    return newComponent.id;
  };

  // Once we have stable IDs, this can be simplified
  processRows = (response: Partial<CanvasResponse>) => {
    const newComponents = response.components;
    const existingKeys = new Set(this.componentsStore.read().keys());
    const rows = response.canvas?.rows;

    if (!rows) return;

    const set = new Set<string>();

    let createdNewComponent = false;
    const isFirstLoad = get(this.firstLoad);

    rows.forEach((row, rowIndex) => {
      const items = row.items ?? [];

      items.forEach((item, columnIndex) => {
        const componentName = item.component;

        if (!componentName) return;

        set.add(componentName ?? "");

        const newResource = newComponents?.[componentName];
        if (!newResource) {
          throw new Error("No component found: " + componentName);
        }

        const newType = (newResource.component?.state?.validSpec?.renderer ??
          (this.allowUnvalidatedSpec
            ? newResource.component?.spec?.renderer
            : undefined)) as CanvasComponentType;
        const existingClass =
          this.componentsStore.getNonReactive(componentName);
        const path = constructPath(rowIndex, columnIndex, newType);

        if (existingClass && areSameType(newType, existingClass.type)) {
          existingClass.update(newResource, path);
        } else {
          createdNewComponent = true;
          this.componentsStore.set(
            componentName,
            createComponent(newResource, this, path),
          );
        }
      });
    });

    const didUpdateRowCount = this._rows.updateFromCanvasRows(rows);

    existingKeys.difference(set).forEach((componentName) => {
      const component = this.componentsStore.getNonReactive(componentName);
      if (component) {
        this.componentsStore.delete(componentName);
      }
    });

    // Calling this function triggers the rows to rerender, ensuring they're up to date
    // with the components Map, which is not reactive
    if ((!didUpdateRowCount && createdNewComponent) || isFirstLoad) {
      this._rows.refresh();
      this.firstLoad.set(false);
    }

    this.selectedComponent.update(($) => $);
  };

  generateId = (row: number | undefined, column: number | undefined) => {
    return `${this.name}--component-${row ?? 0}-${column ?? 0}`;
  };

  createOptimisticResource = (options: {
    type: CanvasComponentType;
    row: number;
    column: number;
    metricsViewName: string;
    metricsViewSpec: V1MetricsViewSpec | undefined;
    spec?: ComponentSpec;
  }): V1Resource => {
    const { type, row, column, metricsViewName, metricsViewSpec } = options;

    const spec =
      options.spec ??
      COMPONENT_CLASS_MAP[type].newComponentSpec(
        metricsViewName,
        metricsViewSpec,
      );

    return {
      meta: {
        name: {
          name: this.generateId(row, column),
          kind: ResourceKind.Component,
        },
      },
      component: {
        state: {
          validSpec: {
            renderer: type,
            rendererProperties:
              spec as unknown as V1ComponentSpecRendererProperties,
          },
        },
        spec: {
          renderer: type,
          rendererProperties:
            spec as unknown as V1ComponentSpecRendererProperties,
        },
      },
    };
  };

  setSelectedComponent = (id: string | null) => {
    // Inspector inputs commit their value on blur. Canvas component elements are
    // not focusable, so clicking another component never blurs the focused input,
    // which would otherwise apply the pending edit to the newly-selected component.
    // Blur the active element first so the edit commits against the component that
    // is still selected, before we switch.
    if (id !== get(this.selectedComponent) && typeof document !== "undefined") {
      const active = document.activeElement;
      if (active instanceof HTMLElement) active.blur();
    }
    this.selectedComponent.set(id);
  };

  setActiveComponent = (id: string) => {
    this.activeComponent.set(id);
  };

  clearActiveComponent = () => {
    this.activeComponent.set(null);
  };

  removeComponent = (componentName: string) => {
    this.componentsStore.delete(componentName);
  };
}

export type ComponentPath = [
  "rows",
  number,
  "items",
  number,
  CanvasComponentType,
];

function constructPath(
  row: number,
  column: number,
  type: CanvasComponentType,
): ComponentPath {
  return ["rows", row, "items", column, type];
}

function areSameType(
  newType: CanvasComponentType,
  existingType: CanvasComponentType,
) {
  if (newType === existingType) return true;

  // For chart types, check if they use the same component class
  if (isChartComponentType(existingType) && isChartComponentType(newType)) {
    const cartesian = [
      "bar_chart",
      "line_chart",
      "area_chart",
      "stacked_bar",
      "stacked_bar_normalized",
    ];

    if (cartesian.includes(existingType) && cartesian.includes(newType)) {
      return true;
    }
    return false;

    // FIXME: The below causes a fatal crash through a dependency cycle
    // const newComponent = CHART_CONFIG[newType].component;
    // const existingComponent = CHART_CONFIG[existingType].component;
    // return newComponent.name === existingComponent.name;
  }

  return isTableComponentType(existingType) && isTableComponentType(newType);
}

function getDefaults(defaultPreset: V1CanvasPreset) {
  const defaultSearchParams = new URLSearchParams();

  const resolvedRange = defaultPreset.timeRange;

  if (resolvedRange) {
    defaultSearchParams.set(ExploreStateURLParams.TimeRange, resolvedRange);
  }

  if (
    defaultPreset.comparisonMode ===
    V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME
  ) {
    defaultSearchParams.set(
      ExploreStateURLParams.ComparisonTimeRange,
      getComparisonTypeFromRangeString(defaultPreset.timeRange),
    );
  }

  Object.entries(defaultPreset?.filterExpr ?? {}).forEach(
    ([metricsViewName, { expression }]) => {
      if (expression) {
        const flattened = flattenExpression(expression);
        const urlFormat = getFilterParam(flattened, [], []);

        if (urlFormat) {
          defaultSearchParams.set(
            `${ExploreStateURLParams.Filters}.${metricsViewName}`,
            urlFormat,
          );
        }
      }
    },
  );

  return defaultSearchParams;
}

function setOrDeleteFilterList(
  yaml: ReturnType<typeof parseDocument>,
  key: "pinned" | "required",
  names: string[],
) {
  if (names.length > 0) {
    yaml.setIn(["filters", key], names);
  } else {
    try {
      yaml.deleteIn(["filters", key]);
    } catch {
      // no-op
    }
  }
}

const customKeySort = (
  a: Pair<unknown, unknown>,
  b: Pair<unknown, unknown>,
) => {
  const priorityKeys = ["type", "display_name", "defaults", "filters"];
  const keyA = a.key?.toString();
  const keyB = b.key?.toString();
  const indexA = priorityKeys.indexOf(keyA ?? "");
  const indexB = priorityKeys.indexOf(keyB ?? "");

  if (indexA > -1 && indexB > -1) {
    return indexA - indexB;
  }
  if (indexA > -1) {
    return -1;
  }
  if (indexB > -1) {
    return 1;
  }
  return 0;
};
