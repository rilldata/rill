import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
import type {
  ScaleStore,
  SimpleConfigurationStore,
} from "@rilldata/web-common/components/data-graphic/state/types";
import { getContext } from "svelte";

export function getGraphicContexts() {
  return {
    xScale: getContext(contexts.scale("x")) as ScaleStore,
    yScale: getContext(contexts.scale("y")) as ScaleStore,
    config: getContext(contexts.config) as SimpleConfigurationStore,
  };
}
