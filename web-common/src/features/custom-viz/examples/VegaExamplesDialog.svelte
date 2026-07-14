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
  import { importVegaExampleWithAgent } from "./import-with-ai";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import ExampleCard from "./ExampleCard.svelte";
  import {
    importVegaExample,
    loadGallery,
    type GalleryExample,
  } from "./example-to-component";

  export let open = false;

  const client = useRuntimeClient();

  const { developerChat } = featureFlags;

  let examples: GalleryExample[] = [];
  let loading = false;
  let importing = false;
  let searchValue = "";
  let selectedSubcategory: string | null = null;

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

  $: subcategories = [...new Set(examples.map((e) => e.subcategory))];

  $: filtered = examples.filter((example) => {
    if (selectedSubcategory && example.subcategory !== selectedSubcategory) {
      return false;
    }
    if (!searchValue) return true;
    const needle = searchValue.toLowerCase();
    return (
      example.title.toLowerCase().includes(needle) ||
      (example.description ?? "").toLowerCase().includes(needle) ||
      example.subcategory.toLowerCase().includes(needle)
    );
  });

  function startBlank() {
    open = false;
    void createResourceAndNavigate(client, ResourceKind.Component);
  }

  async function select(example: GalleryExample) {
    if (importing) return;
    importing = true;
    try {
      if ($developerChat) {
        // AI-first import: the example spec goes into the agent prompt, a placeholder
        // file shows while it generates, and navigation happens once the generated
        // component reconciles cleanly. Runs in the background; the chat sidebar
        // opens and streams progress.
        open = false;
        void importVegaExampleWithAgent(client, example, existingNames);
        return;
      }

      // Without AI: deterministic conversion where possible, verbatim otherwise.
      const filePath = await importVegaExample(client, example, existingNames);
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
      <div class="w-48 shrink-0 border-r overflow-y-auto py-2">
        <button
          class="category"
          class:active={selectedSubcategory === null}
          onclick={() => (selectedSubcategory = null)}
        >
          {m.component_examples_all()}
        </button>
        {#each subcategories as subcategory (subcategory)}
          <button
            class="category"
            class:active={selectedSubcategory === subcategory}
            onclick={() => (selectedSubcategory = subcategory)}
          >
            {subcategory}
          </button>
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
              {#each filtered as example (`${example.subcategory}::${example.name}`)}
                <ExampleCard {example} onSelect={select} />
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
  .category {
    @apply w-full text-left px-4 py-1.5 text-xs text-fg-secondary hover:bg-surface-subtle truncate;
  }
  .category.active {
    @apply text-fg-primary font-medium bg-surface-subtle;
  }
</style>
