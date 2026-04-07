<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtimeServiceGetResource } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { mapResource, type DescribeEntry } from "./resource-mappers";

  interface Props {
    resource: V1Resource;
    allResources?: V1Resource[];
    onviewcomponent?: (componentName: string) => void;
  }

  let { resource, allResources = [], onviewcomponent }: Props = $props();

  const runtimeClient = useRuntimeClient();

  // For canvas resources, fetch component resources from refs
  let componentResources: V1Resource[] = $state([]);
  let loading = $state(false);

  $effect(() => {
    if (resource.canvas) {
      fetchComponentResources(resource);
    } else {
      componentResources = [];
      loading = false;
    }
  });

  async function fetchComponentResources(res: V1Resource) {
    const refs = res.meta?.refs ?? [];
    const componentRefs = refs.filter(
      (r) => r.kind === "rill.runtime.v1.Component",
    );
    if (componentRefs.length === 0) {
      componentResources = [];
      loading = false;
      return;
    }

    loading = true;
    const results = await Promise.all(
      componentRefs.map(async (ref) => {
        try {
          const resp = await runtimeServiceGetResource(runtimeClient, {
            name: { kind: ref.kind, name: ref.name },
          });
          return resp.resource;
        } catch {
          return undefined;
        }
      }),
    );
    componentResources = results.filter(
      (r): r is V1Resource => r !== undefined,
    );
    loading = false;
  }

  interface GroupedItems {
    group: string;
    items: DescribeEntry[];
  }

  interface Section {
    name: string;
    ungrouped: DescribeEntry[];
    groups: GroupedItems[];
  }

  let effectiveResources = $derived(
    componentResources.length > 0 ? componentResources : allResources,
  );
  let entries = $derived(mapResource(resource, effectiveResources));

  // Group entries by section, then by group within each section
  let sections = $derived(
    entries.reduce<Section[]>((acc, entry) => {
      let section = acc.find((s) => s.name === entry.section);
      if (!section) {
        section = { name: entry.section, ungrouped: [], groups: [] };
        acc.push(section);
      }

      if (entry.group) {
        let group = section.groups.find((g) => g.group === entry.group);
        if (!group) {
          group = { group: entry.group, items: [] };
          section.groups.push(group);
        }
        group.items.push(entry);
      } else {
        section.ungrouped.push(entry);
      }

      return acc;
    }, []),
  );

  function hasGroups(section: Section): boolean {
    return section.groups.length > 0;
  }
</script>

{#if loading}
  <div class="flex items-center justify-center py-8">
    <p class="text-xs text-fg-secondary">Loading resource details…</p>
  </div>
{:else}
  <div class="flex flex-col gap-3">
    {#each sections as section (section.name)}
      {#if hasGroups(section)}
        <!-- Collapsible section for grouped items (measures, dimensions, etc.) -->
        <details class="border rounded-md overflow-hidden group" open>
          <summary
            class="px-3 py-1.5 bg-surface-subtle border-b text-xs font-semibold uppercase text-fg-secondary flex items-center justify-between cursor-pointer select-none list-none"
          >
            <span>{section.name}</span>
            <span class="text-fg-muted font-normal normal-case"
              >{section.groups.length} items</span
            >
          </summary>

          <div class="px-3 py-2 flex flex-col gap-y-0.5">
            {#each section.groups as group (group.group)}
              <details class="border rounded group">
                <summary
                  class="px-2.5 py-1.5 text-xs font-medium text-fg-primary cursor-pointer select-none list-none flex items-center justify-between"
                >
                  <span>{group.group}</span>
                  <span class="flex items-center gap-x-1.5">
                    {#if section.name === "Components" && onviewcomponent}
                      <button
                        class="text-primary-500 hover:text-primary-600 text-xs font-normal"
                        onclick={(e) => {
                          e.stopPropagation();
                          onviewcomponent?.(group.group);
                        }}
                      >
                        View
                      </button>
                    {/if}
                    <svg
                      aria-hidden="true"
                      class="w-3 h-3 text-fg-muted transition-transform group-open:rotate-90"
                      viewBox="0 0 16 16"
                      fill="currentColor"
                    >
                      <path
                        d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06Z"
                      />
                    </svg>
                  </span>
                </summary>
                <div class="px-2.5 py-1.5 border-t flex flex-col gap-y-0.5">
                  {#each group.items as item (item.label)}
                    {#if item.label === "Name" || item.label === "Display Name"}
                      <!-- Already shown as summary -->
                    {:else if item.multiline}
                      <div class="flex flex-col gap-1 text-xs">
                        <span class="text-fg-secondary">{item.label}</span>
                        <pre
                          class="bg-surface-subtle rounded p-2 text-xs font-mono overflow-x-auto whitespace-pre-wrap max-h-40">{item.value}</pre>
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
              </details>
            {/each}
          </div>
        </details>
      {:else}
        <!-- Non-collapsible section for flat key-value entries -->
        <div class="border rounded-md overflow-hidden">
          <div
            class="px-3 py-1.5 bg-surface-subtle border-b text-xs font-semibold uppercase text-fg-secondary"
          >
            {section.name}
          </div>
          <div class="px-3 py-2 flex flex-col gap-y-1.5">
            {#each section.ungrouped as item (item.label)}
              {#if item.multiline}
                <div class="flex flex-col gap-1 text-xs">
                  <span class="text-fg-secondary">{item.label}</span>
                  <pre
                    class="bg-surface-subtle rounded p-2 text-xs font-mono overflow-x-auto whitespace-pre-wrap max-h-60">{item.value}</pre>
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
      {/if}
    {/each}

    {#if entries.length === 0}
      <p class="text-xs text-fg-secondary">No spec data available</p>
    {/if}
  </div>
{/if}

<style>
  summary::-webkit-details-marker {
    display: none;
  }
  summary {
    list-style: none;
  }
</style>
