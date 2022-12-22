<!-- @component
A simple composable container for SVG-based data graphics.
-->
<script lang="ts">
  import { GraphicContext, SimpleSVGContainer } from "./index";

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

  export let xMinTweenProps = { duration: 0 };
  export let xMaxTweenProps = { duration: 0 };
  export let yMinTweenProps = { duration: 0 };
  export let yMaxTweenProps = { duration: 0 };

  export let shareXScale = true;
  export let shareYScale = true;

  export let mouseoverValue = undefined;
  export let hovered = undefined;

  /** this makes a wide variety of normal events, such as on:click, available
   * to the consumer
   */
  // const forwardAll = forwardEvents(getComponent(), []);
</script>

<div>
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
    {xMinTweenProps}
    {xMaxTweenProps}
    {yMinTweenProps}
    {yMaxTweenProps}
  >
    <SimpleSVGContainer
      bind:mouseoverValue
      bind:hovered
      let:xScale
      let:yScale
      let:config
      on:scrub-start
      on:scrub-move
      on:scrub-end
    >
      <slot {xScale} {yScale} {mouseoverValue} {config} {hovered} />
    </SimpleSVGContainer>
  </GraphicContext>
</div>
