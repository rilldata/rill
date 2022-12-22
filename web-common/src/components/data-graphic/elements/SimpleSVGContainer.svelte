<!-- @component
A convenience container that creates an svg element and slots in scales, mouseover values, and the configuration object
to the props.
-->
<script lang="ts">
  import { getContext } from "svelte";

  import { mousePositionToDomainActionFactory } from "../actions/mouse-position-to-domain-action-factory";
  import { contexts } from "../constants";
  import type { ScaleStore, SimpleConfigurationStore } from "../state/types";

  const config = getContext(contexts.config) as SimpleConfigurationStore;
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;
  const { coordinates, mousePositionToDomain } =
    mousePositionToDomainActionFactory();

  export let mouseoverValue = undefined;

  $: mouseoverValue = $coordinates;
</script>

<svg use:mousePositionToDomain width={$config.width} height={$config.height}>
  <slot
    config={$config}
    xScale={$xScale}
    yScale={$yScale}
    {mouseoverValue}
    hovered={$coordinates.x !== undefined}
  />
</svg>
