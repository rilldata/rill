<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import {
    type Annotation,
    createAnnotationGroups,
  } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import type { ScaleStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { getAttrs, builderActions } from "bits-ui";
  import { Diamond } from "lucide-svelte";
  import { getContext } from "svelte";

  export let annotations: Annotation[];

  const xScale = getContext(contexts.scale("x")); // as ScaleStore

  $: annotationGroups = createAnnotationGroups(annotations, $xScale);

  // Used to show/hide popover. While tooltip component might be better, popover's base design matches a lot better.
  // TODO: use stack to handle overlapping annotations
  let hoveredAnnotation = -1;
</script>

<div class="relative">
  {#each annotationGroups as annotationGroup, i (i)}
    <Popover.Root open={i === hoveredAnnotation}>
      <Popover.Trigger asChild let:builder>
        <button
          class="absolute bottom-0 w-2 h-4"
          style="left: {annotationGroup.left}px"
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
          on:mouseenter={() => (hoveredAnnotation = i)}
          on:mouseleave={() => (hoveredAnnotation = -1)}
        >
          <Diamond size={10} />
        </button>
      </Popover.Trigger>
      <Popover.Content side="right" sideOffset={20}>
        <div class="flex flex-col gap-y-2">
          {#each annotationGroup.items as annotation, j (j)}
            <div class="flex flex-col">
              <div class="text-popover-foreground">
                {annotation.description}
              </div>
              <div class="text-muted-foreground">
                {annotation.time.toISOString()}
              </div>
            </div>
          {/each}
        </div>
      </Popover.Content>
    </Popover.Root>
  {/each}
</div>
