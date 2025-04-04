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
  type Writable,
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
import { COLUMN_COUNT } from "../layout-util";

export interface LayoutRow {
  itemIds: Writable<string[]>;
  height: Writable<number>;
  itemWidths: Writable<number[]>;
}

export class CanvasEntity {
  name: string;
  components = new Map<string, BaseCanvasComponent>();

  _rows: Writable<LayoutRow[]> = writable([]);

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

  // Once we have stable IDs, this can be simplified
  processRows = (canvasData: Partial<CanvasResponse>) => {
    const newComponents = canvasData.components;
    const existingKeys = new Set(this.components.keys());
    const rows = canvasData.canvas?.rows ?? [];

    const set = new Set<string>();

    const existingRows = get(this._rows);

    let updatedRows = false;
    let createdNewComponent = false;

    if (rows.length < existingRows.length) {
      updatedRows = true;
      this._rows.update((current) => {
        return current.slice(0, rows.length);
      });
    }

    rows.forEach((row, rowIndex) => {
      const items = row.items ?? [];
      const itemIds = items.map((item) => item.component ?? "");
      const existingRow = existingRows[rowIndex];

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

      if (existingRow) {
        existingRow.height.set(row.height ?? 0);
        const existingItemIds = get(existingRow.itemIds);

        if (
          existingItemIds.length !== itemIds.length ||
          itemIds.some((itemId, index) => itemId !== existingItemIds[index])
        ) {
          existingRow.itemIds.set(itemIds);
        }
        existingRow.itemWidths.set(
          items.map((item) => {
            return item.width ?? COLUMN_COUNT / items.length;
          }),
        );
      } else {
        const height = writable(row.height ?? 0);
        this._rows.update((existing) => {
          existing[rowIndex] = {
            itemIds: writable(itemIds),
            height,
            itemWidths: writable(
              items.map((item) => {
                return item.width ?? COLUMN_COUNT / items.length;
              }),
            ),
          };
          return existing;
        });
      }
    });

    existingKeys.difference(set).forEach((componentName) => {
      const component = this.components.get(componentName);
      if (component) {
        this.components.delete(componentName);
      }
    });

    // Only necessary because we are not using stable IDs yet
    if (!updatedRows && createdNewComponent) {
      this._rows.update((r) => r);
    }
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
