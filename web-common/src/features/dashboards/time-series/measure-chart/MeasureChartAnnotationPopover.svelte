<script lang="ts">
  import type { AnnotationGroup } from "./annotation-utils";
  import ThreeDot from "@rilldata/web-common/components/icons/ThreeDot.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import { builderActions, getAttrs } from "bits-ui";

  export let hoveredGroup: AnnotationGroup | null;
  export let onHover: (hovered: boolean) => void;

  const MaxAnnotationCount = 7;
  const PopoverOpenTimeout = 50;

  let open = false;
  let showingMore = false;
  let textHiddenCount = 0;

  $: hasMoreAnnotations =
    !!hoveredGroup &&
    (hoveredGroup.items.length > MaxAnnotationCount || textHiddenCount > 0);

  $: annotationsToShow =
    (showingMore
      ? hoveredGroup?.items
      : hoveredGroup?.items.slice(0, MaxAnnotationCount)) ?? [];

  let lastGroup: AnnotationGroup | null = null;
  let openTimer: ReturnType<typeof setTimeout> | null = null;
  $: handleGroupChange(hoveredGroup);

  function handleGroupChange(group: AnnotationGroup | null) {
    if (lastGroup === group) return;
    lastGroup = group;
    showingMore = false;
    textHiddenCount = 0;
    open = false;
    if (openTimer) {
      clearTimeout(openTimer);
      openTimer = null;
    }
    if (group) {
      openTimer = setTimeout(() => {
        open = true;
        openTimer = null;
      }, PopoverOpenTimeout);
    }
  }

  function checkTextHidden(node: HTMLElement) {
    const check = () => {
      const hidden =
        node.scrollWidth > node.clientWidth ||
        node.scrollHeight > node.clientHeight;
      return hidden;
    };
    let wasHidden = check();
    if (wasHidden) textHiddenCount++;

    return {
      destroy: () => {
        if (wasHidden) textHiddenCount--;
      },
    };
  }
</script>

{#if hoveredGroup}
  <div class="relative">
    <Popover.Root bind:open onOpenChange={() => (showingMore = false)}>
      <Popover.Trigger asChild let:builder>
        <button
          class="absolute bottom-2 w-0 h-0"
          style="left: {hoveredGroup.left}px;"
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
        ></button>
      </Popover.Trigger>
      <Popover.Content side="right" sideOffset={12} class="w-80" padding="0">
        <div
          class="flex flex-col gap-y-1 p-2 max-h-[600px] overflow-y-auto"
          on:mouseenter={() => onHover(true)}
          on:mouseleave={() => onHover(false)}
          role="menu"
          tabindex="-1"
        >
          {#each annotationsToShow as annotation, i (i)}
            <div class="flex flex-col gap-y-1 p-2">
              <div
                class="text-popover-foreground font-medium text-sm w-full {showingMore
                  ? 'text-wrap break-words'
                  : 'h-5 truncate'}"
                use:checkTextHidden
              >
                {annotation.description}
              </div>
              <div class="text-fg-secondary font-normal text-sm">
                {annotation.formattedTimeOrRange}
              </div>
            </div>
          {/each}
          {#if hasMoreAnnotations && !showingMore}
            <button
              on:click={() => (showingMore = true)}
              class="flex flex-row items-center gap-x-1 mb-1 p-1 text-sm text-fg-secondary hover:bg-popover-accent hover:rounded-sm outline-0"
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
