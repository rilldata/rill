<script lang="ts">
  import { forEachRowsList } from "@rilldata/web-common/features/canvas/layout-util";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type {
    V1CanvasItem,
    V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { Writable } from "svelte/store";

  // Name of the component whose usage is shown.
  export let componentName: string;
  export let argsStore: Writable<Record<string, unknown>>;

  const client = useRuntimeClient();

  $: canvasesQuery = useFilteredResources(client, ResourceKind.Canvas);
  $: usedBy = ($canvasesQuery?.data ?? [])
    .map((res) => ({
      resource: res,
      items: referencingItems(res),
    }))
    .filter(({ items }) => items.length > 0);

  // All items in a canvas that reference this component (including in tab groups and containers).
  function referencingItems(res: V1Resource): V1CanvasItem[] {
    const spec = res.canvas?.state?.validSpec ?? res.canvas?.spec;
    const items: V1CanvasItem[] = [];
    forEachRowsList(spec?.rows ?? [], (rows) => {
      for (const row of rows) {
        for (const item of row.items ?? []) {
          if (item.component === componentName && !item.definedInCanvas) {
            items.push(item);
          }
        }
      }
    });
    return items;
  }

  function copyBindings(item: V1CanvasItem) {
    if (!item.params) return;
    argsStore.update((args) => ({ ...args, ...item.params }));
  }
</script>

<div class="flex flex-col border-t">
  <div
    class="px-5 pt-3 pb-1 text-[11px] font-semibold uppercase text-fg-secondary"
  >
    {m.component_used_by()}
  </div>
  {#if usedBy.length === 0}
    <div class="px-5 pb-3 text-xs text-fg-secondary">
      {m.component_not_used()}
    </div>
  {:else}
    {#each usedBy as { resource, items } (resource.meta?.name?.name)}
      {@const canvasName = resource.meta?.name?.name}
      {@const filePath = resource.meta?.filePaths?.[0]}
      <div class="px-5 py-1.5 flex flex-col gap-y-0.5">
        <div class="flex items-center justify-between gap-x-2">
          <a
            class="text-xs font-medium text-primary-600 hover:text-primary-700 truncate"
            href={filePath ? getFileHref(filePath) : undefined}
          >
            {resource.canvas?.state?.validSpec?.displayName || canvasName}
          </a>
          {#if items.length > 1}
            <span class="text-[11px] text-fg-secondary">×{items.length}</span>
          {/if}
        </div>
        {#each items as item, i (i)}
          {#if item.params && Object.keys(item.params).length > 0}
            <button
              class="text-left text-[11px] text-fg-secondary hover:text-fg-primary w-fit"
              onclick={() => copyBindings(item)}
            >
              {m.component_copy_bindings()}
            </button>
          {/if}
        {/each}
      </div>
    {/each}
  {/if}
</div>
