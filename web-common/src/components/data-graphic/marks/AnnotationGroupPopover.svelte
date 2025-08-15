<script lang="ts">
  import { contexts } from "@rilldata/web-common/components/data-graphic/constants";
  import {
    type AnnotationGroup,
    AnnotationsStore,
  } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import type { SimpleConfigurationStore } from "@rilldata/web-common/components/data-graphic/state/types";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { builderActions, getAttrs } from "bits-ui";
  import { getContext } from "svelte";

  export let annotationsStore: AnnotationsStore;

  const {
    hoveredAnnotationGroup,
    annotationPopoverOpened,
    annotationPopoverHovered,
    annotationPopoverTextHiddenCount,
    textHiddenActions,
  } = annotationsStore;
  const plotConfig = getContext<SimpleConfigurationStore>(contexts.config);

  const MaxAnnotationCount = 7;
  const PopoverOpenTimeout = 50;

  $: popoverLeft = $hoveredAnnotationGroup?.left ?? 0;
  $: popoverOffset = $hoveredAnnotationGroup
    ? 8 +
      Math.min(
        $hoveredAnnotationGroup.right - $hoveredAnnotationGroup.left,
        $plotConfig.plotRight - popoverLeft,
      )
    : 0;

  let showingMore = false;
  $: hasMoreAnnotations =
    !!$hoveredAnnotationGroup &&
    ($hoveredAnnotationGroup.items.length > MaxAnnotationCount ||
      $annotationPopoverTextHiddenCount > 0);
  $: annotationsToShow =
    (showingMore
      ? $hoveredAnnotationGroup?.items
      : $hoveredAnnotationGroup?.items.slice(0, MaxAnnotationCount)) ?? [];

  let lastHoveredGroup: AnnotationGroup | undefined = undefined;
  $: handleAnnotationGroupChange($hoveredAnnotationGroup);

  function handleAnnotationGroupChange(
    hoveredAnnotationGroup: AnnotationGroup | undefined,
  ) {
    if (lastHoveredGroup === hoveredAnnotationGroup) return;
    lastHoveredGroup = hoveredAnnotationGroup;

    showingMore = false;
    annotationPopoverOpened.set(false);
    setTimeout(() => {
      annotationPopoverOpened.set(true);
    }, PopoverOpenTimeout);
  }
</script>

{#if $hoveredAnnotationGroup}
  <div class="relative">
    <Popover.Root
      bind:open={$annotationPopoverOpened}
      onOpenChange={() => (showingMore = false)}
    >
      <Popover.Trigger asChild let:builder>
        <button
          class="absolute bottom-2 w-0 h-0"
          style="left: {popoverLeft}px;"
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
        ></button>
      </Popover.Trigger>
      <Popover.Content
        side="right"
        sideOffset={popoverOffset}
        class="w-80"
        padding="0"
      >
        <div
          class="flex flex-col gap-y-1 p-2 max-h-[600px] overflow-y-auto"
          on:mouseenter={() => annotationPopoverHovered.set(true)}
          on:mouseleave={() => annotationPopoverHovered.set(false)}
          role="menu"
          tabindex="-1"
        >
          {#each annotationsToShow as annotation, i (i)}
            <div class="flex flex-col gap-y-1 p-2">
              <div
                class="text-popover-foreground font-medium text-sm w-full {showingMore
                  ? 'text-wrap break-words'
                  : 'h-5 truncate'}"
                use:textHiddenActions
              >
                {annotation.description}
              </div>
              <div class="text-muted-foreground font-normal text-sm">
                {annotation.formattedTimeOrRange}
              </div>
            </div>
          {/each}
          {#if hasMoreAnnotations && !showingMore}
            <button
              on:click={() => (showingMore = true)}
              class="flex flex-row items-center gap-x-1 mb-1 p-1 text-sm text-muted-foreground hover:bg-accent hover:rounded-sm outline-0"
            >
              <ThreeDot className="rotate-90" size="16px" />
              <span>See more</span>
            </button>
          {/if}
        </div>
      </Popover.Content>
    </Popover.Root>
  </div>
{/if}
