<!-- @component
A simple composable container for SVG-based data graphics.
-->
<script lang="ts">
  import type { DomainCoordinates } from "../constants/types";
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
  export let xType: string | undefined = undefined;
  export let yType: string | undefined = undefined;

  export let overflowHidden = true;

  export let xMin: number | undefined = undefined;
  export let xMax: number | undefined = undefined;
  export let yMin: number | undefined = undefined;
  export let yMax: number | undefined = undefined;

  export let xMinTweenProps = { duration: 0 };
  export let xMaxTweenProps = { duration: 0 };
  export let yMinTweenProps = { duration: 0 };
  export let yMaxTweenProps = { duration: 0 };

  export let shareXScale = true;
  export let shareYScale = true;

  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let hovered = undefined;

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
    let:xScale
    let:yScale
    let:config
    on:scrub-start
    on:scrub-move
    on:scrub-end
    on:click
    on:contextmenu
  >
    <slot {xScale} {yScale} {mouseoverValue} {config} {hovered} />
  </SimpleSVGContainer>
</GraphicContext>
