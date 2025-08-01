<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import {
    AnnotationHeight,
    AnnotationsStore,
    AnnotationWidth,
  } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import type { SimpleConfigurationStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import {
    AnnotationDiamondColor,
    AnnotationHighlightBottomColor,
    AnnotationHighlightColor,
    ScrubBoxColor,
  } from "@rilldata/web-common/features/dashboards/time-series/chart-colors.ts";
  import { Diamond } from "lucide-svelte";
  import { getContext } from "svelte";

  export let annotationsStore: AnnotationsStore;
  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let mouseOverThisChart: boolean;

  const { annotationGroups, hoveredAnnotationGroup, annotationPopoverHovered } =
    annotationsStore;
  const plotConfig = getContext<SimpleConfigurationStore>(contexts.config);

  $: annotationsStore.triggerHoverCheck(
    mouseoverValue,
    mouseOverThisChart,
    $annotationPopoverHovered,
  );

  $: hasRange = $hoveredAnnotationGroup?.hasRange;
  $: rangeXStart = $hoveredAnnotationGroup?.left ?? 0;
  $: rangeXEnd = Math.min(
    $hoveredAnnotationGroup?.right ?? 0,
    $plotConfig.plotRight,
  );
  $: rangeYStart = 0;
  $: rangeYEnd = ($hoveredAnnotationGroup?.bottom ?? 0) - AnnotationHeight / 2;
</script>

{#each $annotationGroups as annotationGroup, i (i)}
  {@const hovered = $hoveredAnnotationGroup === annotationGroup}
  <Diamond
    size={AnnotationWidth}
    x={annotationGroup.left - AnnotationWidth / 2}
    y={annotationGroup.top}
    fill={AnnotationDiamondColor}
    opacity={hovered ? 1 : 0.4}
  />
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
