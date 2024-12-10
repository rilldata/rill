import type { Writable } from "svelte/store";
import type { ComponentInputParam } from "../inspector/types";
import type { ImageSpec } from "./image";
import type { KPISpec } from "./kpi";
import type { MarkdownSpec } from "./markdown";
import type { TableSpec } from "./table";

export interface ComponentCommonProperties {
  position: PositionDef;
  title?: string;
  description?: string;
}

export interface ComponentSize {
  width: number;
  height: number;
}
// The CanvasComponent interface is generic over the spec type.
export interface CanvasComponent<T> {
  /**
   * Svelte writable for the spec
   */
  specStore: Writable<T>;
  /**
   * Minimum allowed size for the the component
   * container on the canvas
   */
  minSize: ComponentSize;
  /**
   * The default size of the container when the component
   * is added to the canvas
   */
  defaultSize: ComponentSize;
  /**
   * The minimum condition needed for the spec to be valid
   * for the given component and to be rendered on the canvas
   */
  isValid(spec: T): boolean;
  /**
   * A map of input params which will be used in the visual
   * UI builder
   */
  inputParams(): Record<keyof T, ComponentInputParam>;
}

// TODO: Make it more human friendly and readable, along
// with perc, relative sizes etc.
export interface PositionDef {
  x: number;
  y: number;
  width: number;
  height: number;
}

export type CanvasComponentInput =
  | MarkdownSpec
  | KPISpec
  | ImageSpec
  | TableSpec;

export const commonOptions: Record<
  keyof ComponentCommonProperties,
  ComponentInputParam
> = {
  title: { type: "string", required: false, showInUI: true, label: "Title" },
  description: {
    type: "string",
    required: false,
    showInUI: true,
    label: "Description",
  },
  position: { type: "string", showInUI: false },
};
