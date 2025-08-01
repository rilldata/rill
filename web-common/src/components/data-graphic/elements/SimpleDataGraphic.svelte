<!-- @component
A simple composable container for SVG-based data graphics.
-->
<script lang="ts">
  import type { DomainCoordinates } from "../constants/types";
  import { ScaleType } from "../state";
  import { GraphicContext, SimpleSVGContainer } from "./index";

  export let top: number | undefined = undefined;
  export let bottom: number | undefined = undefined;
  export let left: number | undefined = undefined;
  export let right: number | undefined = undefined;
  export let bodyBuffer: number | undefined = undefined;
  export let marginBuffer: number | undefined = undefined;
  export let width: number | undefined = undefined;
  export let height: number | undefined = undefined;
  export let fontSize: number | undefined = undefined;
  export let textGap: number | undefined = undefined;
  export let xType: ScaleType = ScaleType.DATE;
  export let yType: ScaleType = ScaleType.NUMBER;

  export let overflowHidden = true;

  export let xMin: number | Date | undefined = undefined;
  export let xMax: number | Date | undefined = undefined;
  export let yMin: number | Date | undefined = undefined;
  export let yMax: number | Date | undefined = undefined;

  export let xMinTweenProps = { duration: 0 };
  export let xMaxTweenProps = { duration: 0 };
  export let yMinTweenProps = { duration: 0 };
  export let yMaxTweenProps = { duration: 0 };

  export let shareXScale = true;
  export let shareYScale = true;

  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let hovered = false;

  let mouseOverThisChart = false;

  /** this makes a wide variety of normal events, such as on:click, available
   * to the consumer
   */
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
  {shareXScale}
  {shareYScale}
  {xMinTweenProps}
  {xMaxTweenProps}
  {yMinTweenProps}
  {yMaxTweenProps}
>
  <SimpleSVGContainer
    {overflowHidden}
    bind:mouseoverValue
    bind:hovered
    bind:mouseOverThisChart
    let:xScale
    let:yScale
    let:config
    on:scrub-start
    on:scrub-move
    on:scrub-end
    on:click
    on:contextmenu
  >
    <slot
      {xScale}
      {yScale}
      {mouseoverValue}
      {config}
      {hovered}
      {mouseOverThisChart}
    />
  </SimpleSVGContainer>
</GraphicContext>
