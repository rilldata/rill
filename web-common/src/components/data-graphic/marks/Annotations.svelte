<script lang="ts">
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import {
    AnnotationHeight,
    AnnotationsStore,
  } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import {
    AnnotationHighlightBottomColor,
    AnnotationHighlightColor,
    ScrubBoxColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors.ts";
  import { Diamond } from "lucide-svelte";

  export let annotationsStore: AnnotationsStore;
  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let mouseOverThisChart: boolean;

  const { annotationGroups, hoveredAnnotationGroup, annotationPopoverHovered } =
    annotationsStore;

  $: annotationsStore.triggerHoverCheck(
    mouseoverValue,
    mouseOverThisChart,
    $annotationPopoverHovered,
  );

  $: hasRange = $hoveredAnnotationGroup?.hasRange;
  $: rangeXStart = $hoveredAnnotationGroup?.left ?? 0;
  $: rangeXEnd = $hoveredAnnotationGroup?.right ?? 0;
  $: rangeYStart = 0;
  $: rangeYEnd = ($hoveredAnnotationGroup?.bottom ?? 0) - AnnotationHeight / 2;
</script>

{#each $annotationGroups as annotationGroup, i (i)}
  <Diamond size={10} x={annotationGroup.left} y={annotationGroup.top} />
{/each}

{#if hasRange}
  <g>
    <line
      x1={rangeXStart}
      x2={rangeXStart}
      y1={rangeYStart}
      y2={rangeYEnd}
      stroke={ScrubBoxColor}
      stroke-width={1}
    />
    <line
      x1={rangeXEnd}
      x2={rangeXEnd}
      y1={rangeYStart}
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
      y={rangeYStart}
      width={Math.abs(rangeXStart - rangeXEnd)}
      height={rangeYEnd - rangeYStart}
      fill={AnnotationHighlightColor}
    />
  </g>
{/if}
