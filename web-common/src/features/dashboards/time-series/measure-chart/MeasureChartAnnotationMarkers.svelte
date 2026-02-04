<script lang="ts">
  import type { AnnotationGroup } from "./annotation-utils";
  import { AnnotationWidth, AnnotationHeight } from "./annotation-utils";
  import type { PlotBounds } from "./types";
  import {
    AnnotationDiamondColor,
    AnnotationHighlightColor,
    AnnotationHighlightBottomColor,
    ScrubBoxColor,
  } from "../chart-colors";

  export let groups: AnnotationGroup[];
  export let hoveredGroup: AnnotationGroup | null;
  export let plotBounds: PlotBounds;

  $: hasRange = hoveredGroup?.hasRange ?? false;
  $: rangeXStart = hoveredGroup?.left ?? 0;
  $: rangeXEnd = Math.min(
    hoveredGroup?.right ?? 0,
    plotBounds.left + plotBounds.width,
  );
  $: rangeYEnd = hoveredGroup?.top ?? 0;

  $: halfSize = (AnnotationWidth / 2) * 0.7;
</script>

{#each groups as group (group.index)}
  {@const hovered = hoveredGroup === group}
  {@const cx = group.left}
  {@const cy = group.top + AnnotationHeight / 2}
  <rect
    x={cx - halfSize}
    y={cy - halfSize}
    width={halfSize * 2}
    height={halfSize * 2}
    fill={AnnotationDiamondColor}
    opacity={hovered ? 1 : 0.4}
    transform="rotate(45 {cx} {cy})"
  />
{/each}

{#if hasRange && hoveredGroup}
  <g>
    <line
      x1={rangeXStart}
      x2={rangeXStart}
      y1={plotBounds.top}
      y2={rangeYEnd}
      stroke={ScrubBoxColor}
      stroke-width={1}
    />
    <line
      x1={rangeXEnd}
      x2={rangeXEnd}
      y1={plotBounds.top}
      y2={rangeYEnd}
      stroke={ScrubBoxColor}
      stroke-width={1}
    />
    <line
      x1={rangeXStart}
      x2={rangeXEnd}
      y1={rangeYEnd}
      y2={rangeYEnd}
      stroke={AnnotationHighlightBottomColor}
      stroke-width={2}
    />
  </g>
  <g role="presentation" opacity="0.1">
    <rect
      x={Math.min(rangeXStart, rangeXEnd)}
      y={plotBounds.top}
      width={Math.abs(rangeXStart - rangeXEnd)}
      height={rangeYEnd - plotBounds.top}
      fill={AnnotationHighlightColor}
    />
  </g>
{/if}
