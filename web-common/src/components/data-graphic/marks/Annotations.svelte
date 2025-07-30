<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import type { DomainCoordinates } from "@rilldata/web-common/components/data-graphic/constants/types";
  import { WithBisector } from "@rilldata/web-common/components/data-graphic/functional-components";
  import {
    type Annotation,
    type AnnotationGroup,
    buildLookupTable,
    createAnnotationGroups,
  } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import type {
    ScaleStore,
    SimpleConfigurationStore,
  } from "@rilldata/web-common/components/data-graphic/state/types";
  import { Diamond } from "lucide-svelte";
  import { getContext } from "svelte";

  export let annotations: Annotation[];
  export let mouseoverValue: DomainCoordinates | undefined = undefined;
  export let hovered: boolean;
  export let hoveredAnnotationGroup: AnnotationGroup | undefined;

  const plotConfig: SimpleConfigurationStore = getContext(contexts.config);
  const xScale: ScaleStore = getContext(contexts.scale("x"));

  $: config = $plotConfig;
  $: xScaleFunc = $xScale;

  $: mouseX = mouseoverValue?.xActual;
  $: mouseY = mouseoverValue?.yActual;

  $: annotationGroups = createAnnotationGroups(annotations, xScaleFunc, config);
  $: top = annotationGroups[0]?.top;
  $: lookupTable = buildLookupTable(annotationGroups);

  $: yNearAnnotations = mouseY !== undefined && mouseY > top;
  $: checkHover = hovered && yNearAnnotations && mouseX !== undefined;
  $: hoveredAnnotationGroup = checkHover ? lookupTable[mouseX!] : undefined;
</script>

{#each annotationGroups as annotationGroup, i (i)}
  <Diamond size={10} x={annotationGroup.left} y={annotationGroup.top} />
{/each}
