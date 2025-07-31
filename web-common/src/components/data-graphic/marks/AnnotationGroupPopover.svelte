<script lang="ts">
  import { AnnotationsStore } from "@rilldata/web-common/components/data-graphic/marks/annotations.ts";
  import EllipsisVertical from "@rilldata/web-common/components/icons/EllipsisVertical.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { builderActions, getAttrs } from "bits-ui";

  export let annotationsStore: AnnotationsStore;

  const {
    hoveredAnnotationGroup,
    annotationPopoverOpened,
    annotationPopoverHovered,
  } = annotationsStore;

  const MaxAnnotationCount = 2;
  const PopoverOpenTimeout = 200;

  $: popoverLeft = $hoveredAnnotationGroup?.left ?? 0;
  $: popoverOffset = $hoveredAnnotationGroup
    ? 8 + $hoveredAnnotationGroup.right - $hoveredAnnotationGroup.left
    : 0;

  let showingMore = false;
  $: hasMoreAnnotations =
    !!$hoveredAnnotationGroup &&
    $hoveredAnnotationGroup.items.length > MaxAnnotationCount;
  $: annotationsToShow =
    (showingMore
      ? $hoveredAnnotationGroup?.items
      : $hoveredAnnotationGroup?.items.slice(0, MaxAnnotationCount)) ?? [];

  $: if ($hoveredAnnotationGroup) {
    showingMore = false;
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
        class="w-80 max-h-[600px] overflow-y-auto"
      >
        <div
          class="flex flex-col gap-y-1"
          on:mouseenter={() => annotationPopoverHovered.set(true)}
          on:mouseleave={() => annotationPopoverHovered.set(false)}
          role="menu"
          tabindex="-1"
        >
          {#each annotationsToShow as annotation, i (i)}
            <div class="flex flex-col gap-y-1">
              <div class="text-popover-foreground font-medium text-sm">
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
              class="flex flex-row items-center gap-x-1 mt-1 p-1 text-muted-foreground hover:bg-accent hover:rounded-sm"
            >
              <EllipsisVertical size="16px" />
              <span>See more</span>
            </button>
          {/if}
        </div>
      </Popover.Content>
    </Popover.Root>
  </div>
{/if}
