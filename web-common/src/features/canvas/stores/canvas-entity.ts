import {
  useCanvas,
  type CanvasResponse,
} from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { type V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  derived,
  get,
  writable,
  type Readable,
  type Unsubscriber,
} from "svelte/store";
import { Filters } from "./filters";
import { CanvasResolvedSpec } from "./spec";
import { TimeControls } from "./time-control";
import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
import {
  COMPONENT_CLASS_MAP,
  createComponent,
  isChartComponentType,
  isTableComponentType,
} from "../components/util";
import type { FileArtifact } from "../../entity-management/file-artifact";
import { parseDocument } from "yaml";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import type { CanvasComponentType } from "../components/types";
import { ResourceKind } from "../../entity-management/resource-selectors";
import { Grid } from "./grid";

export class CanvasEntity {
  name: string;
  components = new Map<string, BaseCanvasComponent>();

  _rows: Grid = new Grid(this);

  /**
   * Time controls for the canvas entity containing various
   * time related writables
   */
  timeControls: TimeControls;

  /**
   * Dimension and measure filters for the canvas entity
   */
  filters: Filters;

  /**
   * Spec store containing selectors derived from ResolveCanvas query
   */
  spec: CanvasResolvedSpec;
  selectedComponent = writable<string | null>(null);
  fileArtifact: FileArtifact | undefined;
  parsedContent: Readable<ReturnType<typeof parseDocument>>;
  specStore: CanvasSpecResponseStore;
  firstLoad = true;
  unsubscriber: Unsubscriber;

  constructor(name: string) {
    const instanceId = get(runtime).instanceId;
    this.specStore = useCanvas(instanceId, name, {
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 + 1000 * attemptIndex, 5000),
      queryClient,
    });

    this.name = name;

    this.spec = new CanvasResolvedSpec(this.specStore);
    this.timeControls = new TimeControls(this.specStore);
    this.filters = new Filters(this.spec);

    this.unsubscriber = this.specStore.subscribe((spec) => {
      const filePath = spec.data?.filePath;

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

  unsubscribe = () => {
    // this.unsubscriber();
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
      this.spec.getMetricsViewFromName(metricsViewName),
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
    const rows = canvasData.canvas?.rows ?? [];

    const set = new Set<string>();

    let createdNewComponent = false;

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

    // Only necessary because we are not using stable IDs yet
    if ((!didUpdateRowCount && createdNewComponent) || this.firstLoad) {
      this._rows.refresh();
    }
    this.firstLoad = false;
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
  }) => {
    const { type, row, column, metricsViewName, metricsViewSpec } = options;

    const spec = COMPONENT_CLASS_MAP[type].newComponentSpec(
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
            rendererProperties: spec,
          },
        },
        spec: {
          renderer: type,
          rendererProperties: spec,
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
  return (
    newType === existingType ||
    (isTableComponentType(existingType) && isTableComponentType(newType)) ||
    (isChartComponentType(existingType) && isChartComponentType(newType))
  );
}
