<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { mapResource, type DescribeEntry } from "./resource-mappers";

  export let resource: V1Resource;

  $: entries = mapResource(resource);

  // Group entries by section, preserving insertion order
  $: sections = entries.reduce<{ name: string; items: DescribeEntry[] }[]>(
    (acc, entry) => {
      let section = acc.find((s) => s.name === entry.section);
      if (!section) {
        section = { name: entry.section, items: [] };
        acc.push(section);
      }
      section.items.push(entry);
      return acc;
    },
    [],
  );
</script>

<div class="flex flex-col gap-3">
  {#each sections as section}
    <div class="border rounded-md overflow-hidden">
      <div
        class="px-3 py-1.5 bg-surface-subtle border-b text-xs font-semibold uppercase text-fg-secondary"
      >
        {section.name}
      </div>
      <div class="px-3 py-2 flex flex-col gap-y-1.5">
        {#each section.items as item}
          {#if item.multiline}
            <div class="flex flex-col gap-1 text-xs">
              <span class="text-fg-secondary">{item.label}</span>
              <pre
                class="bg-surface-subtle rounded p-2 text-xs font-mono overflow-x-auto whitespace-pre-wrap max-h-60"
              >{item.value}</pre>
            </div>
          {:else}
            <div class="flex gap-x-4 min-h-[20px] text-xs">
              <span
                class="shrink-0 text-fg-secondary w-36 truncate"
                title={item.label}
              >
                {item.label}
              </span>
              <span
                class="text-fg-primary truncate"
                class:font-mono={item.mono}
                title={item.value}
              >
                {item.value}
              </span>
            </div>
          {/if}
        {/each}
      </div>
    </div>
  {/each}

  {#if entries.length === 0}
    <p class="text-xs text-fg-secondary">No spec data available</p>
  {/if}
</div>
