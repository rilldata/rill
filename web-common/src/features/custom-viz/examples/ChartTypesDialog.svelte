<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { createResourceAndNavigate } from "@rilldata/web-common/features/entity-management/add/new-files";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { navigateToFile } from "@rilldata/web-common/layout/navigation/editor-routing";
  import {
    generatingComponentFilePath,
    importGalleryChartWithAgent,
  } from "./import-with-ai";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExampleCard from "./ExampleCard.svelte";
  import {
    importGalleryChart,
    loadGallery,
    type GalleryEntry,
  } from "./example-to-component";

  export let open = false;

  const client = useRuntimeClient();

  const { developerChat } = featureFlags;

  let examples: GalleryEntry[] = [];
  let loading = false;
  let importing = false;
  let searchValue = "";
  // Sidebar selection: the whole gallery, one Flint category, or one chart type within it.
  let selection: { category: string; chartType?: string } | null = null;

  // An AI import runs in the background after the dialog closes, so reopening the
  // gallery must not start a second one alongside it.
  $: busy = importing || $generatingComponentFilePath !== null;

  $: if (open && !examples.length && !loading) {
    loading = true;
    void loadGallery()
      .then((gallery) => (examples = gallery.examples))
      .finally(() => (loading = false));
  }

  $: componentsQuery = useFilteredResources(client, ResourceKind.Component);
  $: existingNames = ($componentsQuery?.data ?? [])
    .map((res) => res.meta?.name?.name)
    .filter((name): name is string => !!name);

  // Flint groups its chart types into a handful of broad families, which is too coarse on its own
  // for 30-odd types: the sidebar lists each chart type under its family so a reader can jump
  // straight to "Waterfall Chart" instead of scanning all of "Bars".
  $: navigation = buildNavigation(examples);

  function buildNavigation(entries: GalleryEntry[]) {
    const categories = new Map<
      string,
      { name: string; count: number; chartTypes: Map<string, number> }
    >();

    for (const entry of entries) {
      let category = categories.get(entry.category);
      if (!category) {
        category = { name: entry.category, count: 0, chartTypes: new Map() };
        categories.set(entry.category, category);
      }
      category.count++;
      category.chartTypes.set(
        entry.chartType,
        (category.chartTypes.get(entry.chartType) ?? 0) + 1,
      );
    }

    return [...categories.values()].map((category) => ({
      ...category,
      chartTypes: [...category.chartTypes.entries()]
        .map(([name, count]) => ({ name, count }))
        .sort((a, b) => a.name.localeCompare(b.name)),
    }));
  }

  $: matchesSelection = (example: GalleryEntry) => {
    if (!selection) return true;
    if (example.category !== selection.category) return false;
    return !selection.chartType || example.chartType === selection.chartType;
  };

  $: filtered = examples.filter((example) => {
    if (!matchesSelection(example)) return false;
    if (!searchValue) return true;
    const needle = searchValue.toLowerCase();
    return (
      example.chartType.toLowerCase().includes(needle) ||
      example.description.toLowerCase().includes(needle) ||
      example.title.toLowerCase().includes(needle) ||
      example.category.toLowerCase().includes(needle) ||
      example.tags.some((tag) => tag.toLowerCase().includes(needle))
    );
  });

  function startBlank() {
    open = false;
    void createResourceAndNavigate(client, ResourceKind.Component);
  }

  async function select(entry: GalleryEntry) {
    if (busy) return;
    importing = true;
    try {
      if ($developerChat) {
        // AI-first import: the chart type and its channels go into the agent prompt, a
        // placeholder file shows while it generates, and navigation happens once the
        // generated component reconciles cleanly. Runs in the background; the chat
        // sidebar opens and streams progress.
        open = false;
        void importGalleryChartWithAgent(client, entry, existingNames);
        return;
      }

      const filePath = await importGalleryChart(client, entry, existingNames);
      open = false;
      eventBus.emit("notification", {
        message: m.component_example_imported({ path: filePath }),
      });
      await navigateToFile(filePath);
    } catch (error) {
      eventBus.emit("notification", {
        message: extractErrorMessage(error),
        type: "error",
      });
    } finally {
      importing = false;
    }
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-4xl h-[80vh] flex flex-col gap-y-0 p-0">
    <Dialog.Title class="px-6 pt-5 pb-3 border-b">
      <!-- mr-8 keeps the button clear of the dialog's close (X) control. -->
      <div class="flex items-start justify-between gap-x-4 mr-8">
        <div>
          {m.component_examples_title()}
          <div class="pt-1 text-xs font-normal text-fg-secondary">
            {m.component_examples_description()}
          </div>
        </div>
        <Button type="secondary" onClick={startBlank}>
          {m.component_examples_start_blank()}
        </Button>
      </div>
    </Dialog.Title>

    <div class="flex flex-1 min-h-0">
      <div class="w-56 shrink-0 border-r overflow-y-auto py-2">
        <button
          class="category"
          class:active={selection === null}
          onclick={() => (selection = null)}
        >
          <span class="truncate">{m.component_examples_all()}</span>
          <span class="count">{examples.length}</span>
        </button>
        {#each navigation as category (category.name)}
          <button
            class="category mt-1"
            class:active={selection?.category === category.name &&
              !selection?.chartType}
            onclick={() => (selection = { category: category.name })}
          >
            <span class="truncate">{category.name}</span>
            <span class="count">{category.count}</span>
          </button>
          {#each category.chartTypes as chartType (chartType.name)}
            <button
              class="chart-type"
              class:active={selection?.chartType === chartType.name}
              onclick={() =>
                (selection = {
                  category: category.name,
                  chartType: chartType.name,
                })}
            >
              <span class="truncate">{chartType.name}</span>
              <span class="count">{chartType.count}</span>
            </button>
          {/each}
        {/each}
      </div>

      <div class="flex-1 min-w-0 flex flex-col">
        <div class="p-3 border-b">
          <Search bind:value={searchValue} autofocus />
        </div>
        <div class="flex-1 overflow-y-auto p-3">
          {#if loading}
            <div class="text-sm text-fg-secondary p-4">
              {m.common_loading()}
            </div>
          {:else}
            <div class="grid grid-cols-3 gap-3">
              {#each filtered as example (example.name)}
                <ExampleCard {example} onSelect={select} disabled={busy} />
              {:else}
                <div class="text-sm text-fg-secondary p-4 col-span-3">
                  {m.component_examples_empty()}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    </div>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .category,
  .chart-type {
    @apply w-full flex items-center gap-x-2 text-left hover:bg-surface-subtle;
  }
  .category {
    @apply px-4 py-1.5 text-xs font-medium text-fg-primary;
  }
  .chart-type {
    @apply pl-7 pr-4 py-1 text-xs text-fg-secondary;
  }
  .category.active,
  .chart-type.active {
    @apply text-fg-primary font-medium bg-surface-subtle;
  }
  .count {
    @apply ml-auto shrink-0 text-[10px] tabular-nums text-fg-secondary;
  }
</style>
