<script lang="ts">
  import type { AnnotationGroup } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import { ScrubBoxColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors.ts";

  export let annotationGroup: AnnotationGroup;

  const strokeWidth = 1;

  $: xStart = annotationGroup.left;
  $: xEnd = annotationGroup.right;
  $: yStart = 0;
  $: yEnd = annotationGroup.bottom;
</script>

<g>
  <line
    x1={xStart}
    x2={xStart}
    y1={yStart}
    y2={yEnd}
    stroke={ScrubBoxColor}
    stroke-width={strokeWidth}
  />
  <line
    x1={xEnd}
    x2={xEnd}
    y1={yStart}
    y2={yEnd}
    stroke={ScrubBoxColor}
    stroke-width={strokeWidth}
  />
</g>
<g role="presentation" opacity="0.1">
  <rect
    x={Math.min(xStart, xEnd)}
    y={yStart}
    width={Math.abs(xStart - xEnd)}
    height={yEnd - yStart}
  />
</g>
