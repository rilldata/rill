import type { ScaleLinear, ScaleTime } from "d3-scale";
import type { Readable } from "svelte/store";

export interface ScaleStore extends Readable<ScaleLinear<number, number> | ScaleTime<Date, number>> {
  type: string
};

export interface SimpleDataGraphicConfigurationArguments {
  width: number;
  height: number;
  left: number;
  right: number;
  top: number;
  bottom: number;
  fontSize: number;
  textGap: number;
  xType: string; // FIXME: we should have an enum here
  yType: string; // FIXME: we should have an enum here
  xMin: (number | Date);
  xMax: (number | Date);
  yMin: (number | Date);
  yMax: (number | Date);
  bodyBuffer: number;
  marginBuffer: number;
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

export interface SimpleDataGraphicConfiguration extends SimpleDataGraphicConfigurationArguments, SimpleGraphicConfigurationDerivations { };

export interface CascadingContextStore<Arguments, StateStructure> extends Readable<StateStructure> {
  hasParentCascade: boolean;
  reconcileProps: (props: Arguments) => void
}

export type SimpleConfigurationStore = CascadingContextStore<SimpleDataGraphicConfigurationArguments, SimpleDataGraphicConfiguration>;