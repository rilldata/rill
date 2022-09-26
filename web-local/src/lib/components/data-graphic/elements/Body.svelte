<!--
@component
Creates a clip path group element that hides parts of children out of the viewport range.

Optionally allows borders (see the `border` & associated props) & a background (see the `bg` & associated props).
-->
<script lang="ts">
  import { getContext } from "svelte";
  import { contexts } from "../constants";
  import type { SimpleConfigurationStore } from "../state/types";

  /** By default, we clip any child outside of the body bounds. */
  export let clipOutsideBounds = true;

  /** Background
   *  ----------
   * A Body element has an optional background.
   */
  export let bg = false;
  export let bgColor = "rgb(200,200,200)";
  export let bgOpacity = 1;

  /** Border
   *  ------
   * A Body element can have a border on each side.
   * You can either activate all borders & change all their props
   * or select individual borders.
   */

  export let border = false;
  export let leftBorder = border;
  export let rightBorder = border;
  export let topBorder = border;
  export let bottomBorder = border;

  export let borderSize = 1;
  export let leftBorderSize = borderSize;
  export let rightBorderSize = borderSize;
  export let topBorderSize = borderSize;
  export let bottomBorderSize = borderSize;

  export let borderColor = "lightgray";
  export let leftBorderColor = borderColor;
  export let rightBorderColor = borderColor;
  export let topBorderColor = borderColor;
  export let bottomBorderColor = borderColor;

  const config = getContext(contexts.config) as SimpleConfigurationStore;
</script>

<clipPath id="data-graphic-{$config.id}">
  <rect
    x={$config.bodyLeft}
    y={$config.bodyTop}
    width={$config.graphicWidth}
    height={$config.graphicHeight}
  />
</clipPath>

{#if bg}
  <rect
    fill={bgColor}
    opacity={bgOpacity}
    x={$config.bodyLeft}
    y={$config.bodyTop}
    width={$config.graphicWidth}
    height={$config.graphicHeight}
  />
{/if}

<g
  clip-path={clipOutsideBounds ? `url(#data-graphic-${$config.id})` : undefined}
>
  <slot />
</g>
<g>
  {#if leftBorder}
    <line
      x1={$config.plotLeft}
      x2={$config.plotLeft}
      y1={$config.plotTop}
      y2={$config.plotBottom}
      stroke-width={leftBorderSize}
      stroke={leftBorderColor}
    />
  {/if}
  {#if rightBorder}
    <line
      x1={$config.plotRight}
      x2={$config.plotRight}
      y1={$config.plotTop}
      y2={$config.plotBottom}
      stroke-width={rightBorderSize}
      stroke={rightBorderColor}
    />
  {/if}
  {#if topBorder}
    <line
      x1={$config.plotLeft}
      x2={$config.plotRight}
      y1={$config.plotTop}
      y2={$config.plotTop}
      stroke-width={topBorderSize}
      stroke={topBorderColor}
    />
  {/if}
  {#if bottomBorder}
    <line
      x1={$config.plotLeft}
      x2={$config.plotRight}
      y1={$config.plotBottom}
      y2={$config.plotBottom}
      stroke-width={bottomBorderSize}
      stroke={bottomBorderColor}
    />
  {/if}
</g>
