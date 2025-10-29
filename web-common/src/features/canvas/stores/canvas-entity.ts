import { goto } from "$app/navigation";
import { page } from "$app/stores";
import {
  useCanvas,
  type CanvasResponse,
} from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import {
  defaultPrimaryColors,
  defaultSecondaryColors,
} from "@rilldata/web-common/features/themes/color-config";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetResourceQueryKey,
  runtimeServiceGetResource,
  type V1CanvasSpec,
  type V1ComponentSpecRendererProperties,
  type V1MetricsViewSpec,
  type V1Resource,
  type V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import chroma, { type Color } from "chroma-js";
import {
  derived,
  get,
  writable,
  type Readable,
  type Unsubscriber,
  type Writable,
} from "svelte/store";
import { parseDocument } from "yaml";
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
import { Filters } from "./filters";
import { Grid } from "./grid";
import { TimeControls } from "./time-control";

// Store for managing URL search parameters
// Which may be in the URL or in the Canvas YAML
// Set returns a boolean indicating whether the value was set
export type SearchParamsStore = {
  subscribe: (run: (value: URLSearchParams) => void) => Unsubscriber;
  set: (key: string, value?: string, checkIfSet?: boolean) => boolean;
  clearAll: () => void;
};

export class CanvasEntity {
  name: string;
  components = new Map<string, BaseCanvasComponent>();

  _rows: Grid = new Grid(this);

  // Time state controls
  timeControls: TimeControls;

  // Dimension and measure filter state
  filters: Filters;

  // Metrics view selectors
  metricsView: MetricsViewSelectors;

  // Canvas resource infered from YAML spec
  spec: Readable<V1CanvasSpec | undefined>;
  selectedComponent = writable<string | null>(null);
  fileArtifact: FileArtifact | undefined;
  parsedContent: Readable<ReturnType<typeof parseDocument>>;
  specStore: CanvasSpecResponseStore;
  // Tracks whether the canvas been loaded (and rows processed) for the first time
  firstLoad = writable(true);
  theme: Writable<{ primary?: Color; secondary?: Color }> = writable({});
  themeSpec: Writable<V1ThemeSpec | undefined> = writable(undefined);
  unsubscriber: Unsubscriber;
  lastVisitedState: Writable<string | null> = writable(null);

  constructor(
    name: string,
    private instanceId: string,
  ) {
    this.specStore = useCanvas(
      instanceId,
      name,
      {
        retry: 3,
        retryDelay: (attemptIndex) =>
          Math.min(1000 + 1000 * attemptIndex, 5000),
      },
      queryClient,
    );

    this.name = name;

    const searchParamsStore: SearchParamsStore = (() => {
      return {
        subscribe: derived(page, ({ url: { searchParams } }) => searchParams)
          .subscribe,
        set: (key: string, value: string | undefined, checkIfSet = false) => {
          const url = get(page).url;

          if (checkIfSet && url.searchParams.has(key)) return false;

          if (value === undefined || value === null || value === "") {
            url.searchParams.delete(key);
          } else {
            url.searchParams.set(key, value);
          }
          goto(url.toString(), { replaceState: true }).catch(console.error);
          return true;
        },
        clearAll: () => {
          const url = get(page).url;
          url.searchParams.forEach((_, key) => {
            url.searchParams.delete(key);
          });

          goto(url.toString(), { replaceState: true }).catch(console.error);
        },
      };
    })();

    this.spec = derived(this.specStore, ($specStore) => {
      return $specStore.data?.canvas;
    });

    this.metricsView = new MetricsViewSelectors(
      instanceId,
      derived(this.specStore, ($specStore) => {
        return $specStore.data?.metricsViews || {};
      }),
    );

    this.timeControls = new TimeControls(
      this.specStore,
      searchParamsStore,
      undefined,
      this.name,
    );
    this.filters = new Filters(this.metricsView, searchParamsStore);

    searchParamsStore.subscribe((searchParams) => {
      const themeFromUrl = searchParams.get("theme");
      if (themeFromUrl) {
        this.processTheme(themeFromUrl, undefined).catch(console.error);
      }
    });

    this.unsubscriber = this.specStore.subscribe((spec) => {
      const filePath = spec.data?.filePath;
      const theme = spec.data?.canvas?.theme;
      const embeddedTheme = spec.data?.canvas?.embeddedTheme;

      this.processTheme(theme, embeddedTheme).catch(console.error);

      if (!filePath) {
        return;
      }

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

      if (spec.data) {
        this.processRows(spec.data);
      }
    });
  }

  // Not currently being used
  unsubscribe = () => {
    // this.unsubscriber();
  };

  saveSnapshot = (filterState: string) => {
    this.lastVisitedState.set(filterState);
  };

  restoreSnapshot = async () => {
    const lastVisitedState = get(this.lastVisitedState);

    if (lastVisitedState) {
      await goto(`?${lastVisitedState}`, {
        replaceState: true,
      });
    }
  };

  duplicateItem = (id: string) => {
    const component = this.components.get(id);
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
  processRows = (canvasData: Partial<CanvasResponse>) => {
    const newComponents = canvasData.components;
    const existingKeys = new Set(this.components.keys());
    const rows = canvasData.canvas?.rows;

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

        const newType = newResource.component?.state?.validSpec
          ?.renderer as CanvasComponentType;
        const existingClass = this.components.get(componentName);
        const path = constructPath(rowIndex, columnIndex, newType);

        if (existingClass && areSameType(newType, existingClass.type)) {
          existingClass.update(newResource, path);
        } else {
          createdNewComponent = true;
          this.components.set(
            componentName,
            createComponent(newResource, this, path),
          );
        }
      });
    });

    const didUpdateRowCount = this._rows.updateFromCanvasRows(rows);

    existingKeys.difference(set).forEach((componentName) => {
      const component = this.components.get(componentName);
      if (component) {
        this.components.delete(componentName);
      }
    });

    // Calling this function triggers the rows to rerender, ensuring they're up to date
    // with the components Map, which is not reactive
    if ((!didUpdateRowCount && createdNewComponent) || isFirstLoad) {
      this._rows.refresh();
      this.firstLoad.set(false);
    }
  };

  processTheme = async (
    themeName: string | undefined,
    embeddedTheme: V1ThemeSpec | undefined,
  ) => {
    let themeSpec: V1ThemeSpec | undefined;

    if (themeName) {
      const response = await queryClient.fetchQuery({
        queryKey: getRuntimeServiceGetResourceQueryKey(this.instanceId, {
          "name.kind": ResourceKind.Theme,
          "name.name": themeName,
        }),
        queryFn: () =>
          runtimeServiceGetResource(this.instanceId, {
            "name.kind": ResourceKind.Theme,
            "name.name": themeName,
          }),
      });

      themeSpec = response.resource?.theme?.spec;
    } else if (embeddedTheme) {
      themeSpec = embeddedTheme;
    }

    this.themeSpec.set(themeSpec);

    this.theme.set({
      primary: themeSpec?.primaryColorRaw
        ? chroma(themeSpec.primaryColorRaw)
        : chroma(`hsl(${defaultPrimaryColors[500]})`),
      secondary: themeSpec?.secondaryColorRaw
        ? chroma(themeSpec.secondaryColorRaw)
        : chroma(`hsl(${defaultSecondaryColors[500]})`),
    });
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
    this.selectedComponent.set(id);
  };

  removeComponent = (componentName: string) => {
    this.components.delete(componentName);
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
