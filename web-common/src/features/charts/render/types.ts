export type { Spec as VegaSpec, View } from "vega";
export type { VisualizationSpec } from "vega-embed";
export type { TopLevelSpec as VegaLiteSpec } from "vega-lite";

export type SignalListener = (name: string, value: unknown) => void;

export type SignalListeners = {
  [key: string]: SignalListener;
};
