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

  export let mouseoverValues;

  $: mouseoverValues = $coordinates;
</script>

<svg use:mousePositionToDomain width={$config.width} height={$config.height}>
  <slot {config} xScale={$xScale} yScale={$yScale} {mouseoverValues} />
</svg>
