import { createContext } from "@rilldata/web-common/lib/context";

import type { Writable } from "svelte/store";
import type { ScaleLinear } from "d3-scale";
import type { PlotConfig } from "../../utils";

export const dataGraphicContext = {
  x: createContext<Writable<ScaleLinear<number, number>>>(
    "rill:data-graphic:X",
  ),
  y: createContext<Writable<ScaleLinear<number, number>>>(
    "rill:data-graphic:Y",
  ),
  plotConfig: createContext<Writable<PlotConfig>>(
    "rill:data-graphic:plot-config",
  ),
};
