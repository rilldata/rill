<!-- @component
A convenience container that creates an svg element and slots in scales, mouseover values, and the configuration object
to the props.
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { mousePositionToDomainActionFactory } from "../actions/mouse-position-to-domain-action-factory";
  import { createScrubAction } from "../actions/scrub-action-factory";
  import { contexts } from "../constants";
  import type { DomainCoordinates } from "../constants/types";
  import type {
    ScaleStore,
    SimpleConfigurationStore,
  } from "@rilldata/web-common/components/data-graphic/state/types";

  const config = getContext<SimpleConfigurationStore>(contexts.config);
  const xScale = getContext<ScaleStore>(contexts.scale("x"));
  const yScale = getContext<ScaleStore>(contexts.scale("y"));
  const { coordinates, mousePositionToDomain, mouseover } =
    mousePositionToDomainActionFactory();

  const scrubActionObject = createScrubAction({
    plotLeft: $config?.plotLeft,
    plotRight: $config?.plotRight,
    plotTop: $config?.plotTop,
    plotBottom: $config?.plotBottom,
    startEventName: "scrub-start",
    moveEventName: "scrub-move",
    endEventName: "scrub-end",
  });

  // pull out the scrub action to be attached to the svg element
  const scrub = scrubActionObject.scrubAction;
  // const scrubCoordinates = scrubActionObject.coordinates;

  // make sure to reactively update the action store
  $: scrubActionObject.updatePlotBounds({
    plotLeft: $config?.plotLeft,
    plotRight: $config?.plotRight,
    plotTop: $config?.plotTop,
    plotBottom: $config?.plotBottom,
  });

  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let hovered: boolean = false;
  export let mouseOverThisChart: boolean = false;
  export let overflowHidden = true;

  $: mouseoverValue = $coordinates;

  $: hovered = $coordinates.x !== undefined;
  $: mouseOverThisChart = $mouseover;
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<svg
  role="application"
  tabindex="-1"
  style="overflow: {overflowHidden ? 'hidden' : 'visible'}"
  style:outline="none"
  use:scrub
  on:scrub-start
  on:scrub-end
  on:scrub-move
  use:mousePositionToDomain
  on:click
  on:contextmenu
  width={$config.width}
  height={$config.height}
>
  <slot
    config={$config}
    xScale={$xScale}
    yScale={$yScale}
    {mouseoverValue}
    {hovered}
  />
</svg>
