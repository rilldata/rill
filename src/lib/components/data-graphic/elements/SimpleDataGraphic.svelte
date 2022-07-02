<!-- @component
A simple composable container for SVG-based data graphics.
-->
<script lang="ts">
  import GraphicContext from "./elements/GraphicContext.svelte";
  import { mousePositionToDomainActionFactory } from "./actions/mouse-position-to-domain-action-factory";
  import SimpleSVGContainer from "./elements/SimpleSVGContainer.svelte";

  export let top = 40;
  export let bottom = 40;
  export let left = 42;
  export let right = 42;
  export let bodyBuffer = 4;
  export let marginBuffer = 0;
  export let width = 360;
  export let height = 120;
  export let fontSize = 12;
  export let textGap = 4;
  export let xType = undefined;
  export let yType = undefined;

  export let xMin = undefined;
  export let xMax = undefined;
  export let yMin = undefined;
  export let yMax = undefined;

  export let mouseoverValue;

  $: parameters = {
    width,
    height,
    top,
    bottom,
    left,
    right,
    fontSize,
    textGap,
    xType,
    yType,
    xMin,
    xMax,
    yMin,
    yMax,
    bodyBuffer,
    marginBuffer,
  };
</script>

<GraphicContext {...parameters}>
  <SimpleSVGContainer
    bind:mouseoverValue
    let:xScale
    let:yScale
    let:config
    let:hovered
  >
    <slot {xScale} {yScale} {mouseoverValue} {config} {hovered} />
  </SimpleSVGContainer>
</GraphicContext>
