<!-- @component
A simple composable container for SVG-based data graphics.
-->
<script lang="ts">
  import { GraphicContext, SimpleSVGContainer } from "../elements";

  export let top = undefined;
  export let bottom = undefined;
  export let left = undefined;
  export let right = undefined;
  export let bodyBuffer = undefined;
  export let marginBuffer = undefined;
  export let width = undefined;
  export let height = undefined;
  export let fontSize = undefined;
  export let textGap = undefined;
  export let xType = undefined;
  export let yType = undefined;

  export let xMin = undefined;
  export let xMax = undefined;
  export let yMin = undefined;
  export let yMax = undefined;

  export let mouseoverValue = undefined;
</script>

<GraphicContext
  {width}
  {height}
  {top}
  {bottom}
  {left}
  {right}
  {fontSize}
  {textGap}
  {xType}
  {yType}
  {xMin}
  {xMax}
  {yMin}
  {yMax}
  {bodyBuffer}
  {marginBuffer}
>
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
