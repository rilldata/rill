import {
  useCanvas,
  type CanvasResponse,
} from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  type V1MetricsViewSpec,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
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
import { COMPONENT_CLASS_MAP, createComponent } from "../components/util";
import type { FileArtifact } from "../../entity-management/file-artifact";
import { parseDocument } from "yaml";
import { fileArtifacts } from "../../entity-management/file-artifacts";
import type { CanvasComponentType } from "../components/types";
import { ResourceKind } from "../../entity-management/resource-selectors";

export class CanvasEntity {
  name: string;
  _components = writable(new Map<string, BaseCanvasComponent>());

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
    this.unsubscriber();
  };

  // This is not ideal behavior, we don't need to be recreating these components
  // Once we have stable IDs, this can be easily updated.
  processRows = (canvasData: Partial<CanvasResponse>) => {
    const components = canvasData.components;
    const rows = canvasData.canvas?.rows ?? [];

    const newComponents = new Map<string, BaseCanvasComponent>();

    rows.forEach((row, rowIndex) => {
      const items = row.items ?? [];

      items.forEach((item, columnIndex) => {
        const componentName = item.component;
        if (!componentName) return;

        const resource = components?.[componentName];
        if (!resource) {
          throw new Error("No component found: " + componentName);
        }
        const path = constructPath(
          rowIndex,
          columnIndex,
          resource.component?.state?.validSpec?.renderer as CanvasComponentType,
        );

        newComponents.set(componentName, createComponent(resource, this, path));
      });
    });

    this._components.set(newComponents);
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

  initializeOrUpdateComponent = (
    resource: V1Resource,
    row: number,
    column: number,
  ) => {
    const id = resource.meta?.name?.name as string;
    const component = get(this._components).get(id);
    const path = constructPath(
      row,
      column,
      resource.component?.state?.validSpec?.renderer as CanvasComponentType,
    );
    if (component) {
      component.update(resource, path);
    } else {
      this._components.update((components) => {
        const newComponent = createComponent(resource, this, path);
        components.set(id, newComponent);
        return components;
      });
    }
  };

  removeComponent = (componentName: string) => {
    this._components.update((components) => {
      components.delete(componentName);
      return components;
    });
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
