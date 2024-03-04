import equal from "fast-deep-equal";
import type { EmbedOptions, VisualizationSpec } from "vega-embed";
import { vega } from "vega-embed";
import type { SignalListeners, View } from "./types";

export function updateMultipleDatasetsInView(
  view: View,
  data: Record<string, unknown>,
): void {
  for (const [name, value] of Object.entries(data)) {
    const getType = {};
    if (value) {
      if (!!value && getType.toString.call(value) === "[object Function]") {
        const parsedValue = value as (dataset: unknown) => unknown;
        parsedValue(view.data(name));
      } else {
        view.change(
          name,
          vega
            .changeset()
            .remove(() => true)
            .insert(value),
        );
      }
    }
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function shallowEqual<T extends Record<string, any>>(
  a: T = {} as T,
  b: T = {} as T,
  ignore: Set<string> = new Set(),
): boolean {
  const aKeys = Object.keys(a);
  const bKeys = Object.keys(b);

  return (
    a === b ||
    (aKeys.length === bKeys.length &&
      aKeys.filter((k) => !ignore.has(k)).every((key) => a[key] === b[key]))
  );
}

export function removeSignalListenersFromView(
  view: View,
  signalListeners: SignalListeners,
): boolean {
  const signalNames = Object.keys(signalListeners);
  for (const signalName of signalNames) {
    try {
      view.removeSignalListener(signalName, signalListeners[signalName]);
    } catch (error) {
      // eslint-disable-next-line no-console
      console.warn("Cannot remove invalid signal listener.", error);
    }
  }

  return signalNames.length > 0;
}

export function addSignalListenersToView(
  view: View,
  signalListeners: SignalListeners,
): boolean {
  const signalNames = Object.keys(signalListeners);

  for (const signalName of signalNames) {
    try {
      view.addSignalListener(signalName, signalListeners[signalName]);
    } catch (error) {
      console.warn("Cannot add invalid signal listener.", error);
    }
  }

  return signalNames.length > 0;
}

export function getUniqueFieldNames(specs: VisualizationSpec[]): Set<string> {
  return new Set(specs.flatMap((o) => Object.keys(o)));
}

interface SpecChanges {
  width: false | number;
  height: false | number;
  isExpensive: boolean;
}

export function computeSpecChanges(
  newSpec: VisualizationSpec,
  oldSpec: VisualizationSpec,
): false | SpecChanges {
  if (newSpec === oldSpec) return false;

  const changes: SpecChanges = {
    width: false,
    height: false,
    isExpensive: false,
  };

  const hasWidth = "width" in newSpec || "width" in oldSpec;
  const hasHeight = "height" in newSpec || "height" in oldSpec;

  if (
    hasWidth &&
    (!("width" in newSpec) ||
      !("width" in oldSpec) ||
      newSpec.width !== oldSpec.width)
  ) {
    if ("width" in newSpec && typeof newSpec.width === "number") {
      changes.width = newSpec.width;
    } else {
      changes.isExpensive = true;
    }
  }

  if (
    hasHeight &&
    (!("height" in newSpec) ||
      !("height" in oldSpec) ||
      newSpec.height !== oldSpec.height)
  ) {
    if ("height" in newSpec && typeof newSpec.height === "number") {
      changes.height = newSpec.height;
    } else {
      changes.isExpensive = true;
    }
  }

  const fieldNames = [...getUniqueFieldNames([newSpec, oldSpec])].filter(
    (f) => f !== "width" && f !== "height",
  );

  if (
    fieldNames.some(
      (field) =>
        !(field in newSpec) ||
        !(field in oldSpec) ||
        !equal(
          newSpec[field as keyof typeof newSpec],
          oldSpec[field as keyof typeof oldSpec],
        ),
    )
  ) {
    changes.isExpensive = true;
  }

  return changes.width !== false ||
    changes.height !== false ||
    changes.isExpensive
    ? changes
    : false;
}

export function combineSpecWithDimension(
  spec: VisualizationSpec,
  options: EmbedOptions,
): VisualizationSpec {
  const { width, height } = options;
  if (typeof width !== "undefined" && typeof height !== "undefined") {
    return { ...spec, width, height };
  }
  if (typeof width !== "undefined") {
    return { ...spec, width };
  }
  if (typeof height !== "undefined") {
    return { ...spec, height };
  }
  return spec;
}
