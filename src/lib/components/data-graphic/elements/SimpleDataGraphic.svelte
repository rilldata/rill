<!-- @component
A simple composable container for SVG-based data graphics.
-->
<script lang="ts">
  import { GraphicContext, SimpleSVGContainer } from "../elements";
  import { get_current_component as getComponent } from "svelte/internal";
  import { forwardEvents } from "../actions/forward-events-action-factory";

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

  export let shareXScale = true;
  export let shareYScale = true;

  export let mouseoverValue = undefined;

  /** this makes a wide variety of normal events, such as on:click, available
   * to the consumer
   */
  const forwardAll = forwardEvents(getComponent());
</script>

<div use:forwardAll>
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
    {shareXScale}
    {shareYScale}
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
</div>
