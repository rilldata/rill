import type { ScaleLinear, ScaleTime } from "d3-scale";
import type { Readable } from "svelte/store";

export type GraphicScale =
  | ScaleLinear<number, number>
  | ScaleTime<Date, number>;

export interface ScaleStore extends Readable<GraphicScale> {
  type: string;
}

export interface SimpleDataGraphicConfigurationArguments {
  id: string;
  width: number;
  height: number;
  left: number;
  right: number;
  top: number;
  bottom: number;
  fontSize: number;
  textGap: number;
  xType: ScaleType;
  yType: ScaleType;
  xMin: number | Date;
  xMax: number | Date;
  yMin: number | Date;
  yMax: number | Date;
  bodyBuffer: number;
  marginBuffer: number;
  devicePixelRatio: number;
}

export interface SimpleGraphicConfigurationDerivations {
  bodyLeft: number;
  bodyRight: number;
  bodyTop: number;
  bodyBottom: number;
  plotLeft: number;
  plotRight: number;
  plotTop: number;
  plotBottom: number;
  graphicWidth: number;
  graphicHeight: number;
}

export interface SimpleDataGraphicConfiguration
  extends SimpleDataGraphicConfigurationArguments,
    SimpleGraphicConfigurationDerivations {}

export interface CascadingContextStore<Arguments, StateStructure>
  extends Readable<StateStructure> {
  hasParentCascade: boolean;
  reconcileProps: (props: Arguments) => void;
}

export type SimpleConfigurationStore = CascadingContextStore<
  SimpleDataGraphicConfigurationArguments,
  SimpleDataGraphicConfiguration
>;
