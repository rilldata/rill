<script lang="ts">
  import type {
    AdvancedFilter,
    AdvancedFilterMentionOption,
  } from "@rilldata/web-common/features/dashboards/filters/advanced-filters/advanced-filter.ts";
  import {
    autoUpdate,
    computePosition,
    offset,
    flip,
    shift,
    inline,
  } from "@floating-ui/dom";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";

  let {
    items,
    refNode,
    onSelect,
  }: {
    items: AdvancedFilterMentionOption[];
    refNode: HTMLElement;
    onSelect: (ctx: AdvancedFilter) => void;
  } = $props();

  let dimensions = $derived(
    items
      .filter((item) => item.type === "dimension")
      .map((item) => item.dimension),
  );
  let measures = $derived(
    items.filter((item) => item.type === "measure").map((item) => item.measure),
  );

  function positionHandler(node: Node, ref: HTMLElement) {
    if (!(node instanceof HTMLElement)) return;

    let refNode = ref;
    let cleanup: (() => void) | null = null;

    // Temporary minimal implementation of https://github.com/romkor/svelte-portal/blob/master/src/Portal.svelte
    // We wont need this once we upgrade to svelte5 and switch to using dropdown component.
    const update = (newRef: HTMLElement) => {
      cleanup?.();
      document.body.appendChild(node);

      refNode = newRef;
      cleanup = autoUpdate(refNode, node, compute);
      compute();
    };

    const compute = () => {
      void computePosition(refNode, node, {
        placement: "top-start",
        middleware: [offset(10), flip(), shift(), inline()],
      }).then(({ x, y }) => {
        Object.assign(node.style, {
          left: `${x}px`,
          top: `${y}px`,
        });
      });
    };

    update(ref);
    return {
      update,
      destroy() {
        cleanup?.();
        if (node.parentNode) {
          node.parentNode.removeChild(node);
        }
      },
    };
  }

  function selectItem(name: string) {
    onSelect({ name, sql: "" });
  }
</script>

<div class="inline-chat-context-dropdown" use:positionHandler={refNode}>
  <div class="dropdown-content">
    <div>Dimensions</div>
    {#each dimensions as dimension (dimension.name)}
      {@const label = getDimensionDisplayName(dimension)}
      <button
        class="context-item"
        type="button"
        onclick={() => selectItem(dimension.name!)}>{label}</button
      >
    {/each}

    <div class="section-boundary"></div>

    <div>Measures</div>
    {#each measures as measure (measure.name)}
      {@const label = getMeasureDisplayName(measure)}
      <button
        class="context-item"
        type="button"
        onclick={() => selectItem(measure.name!)}>{label}</button
      >
    {/each}
  </div>
</div>

<style lang="postcss">
  .inline-chat-context-dropdown {
    @apply flex flex-col absolute top-0 left-0 p-1.5 z-50;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .dropdown-content {
    @apply flex flex-col w-[400px] max-h-[500px] overflow-auto;
  }

  .section-boundary {
    @apply border-b;
  }

  .context-item {
    @apply flex flex-row items-center gap-x-2 px-2 py-1 w-full;
    @apply cursor-default select-none rounded-sm outline-none;
    @apply text-sm text-left text-wrap break-words;
  }
  .context-item:hover {
    @apply cursor-pointer;
  }
</style>
