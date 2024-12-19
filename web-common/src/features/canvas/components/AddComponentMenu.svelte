<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { ChevronDown, Plus } from "lucide-svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getNameFromFile } from "../../entity-management/entity-mappers";
  // import { featureFlags } from "../feature-flags";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { createResourceFile } from "../../file-explorer/new-files";

  // const { ai } = featureFlags;

  export let addComponent: (componentName: string) => void;

  let open = false;
  let value = "";

  // We want to get only valid components here. Hence using ListResources API
  $: componentsQuery = useClientFilteredResources(
    $runtime.instanceId,
    ResourceKind.Component,
  );
  $: componentFileNames =
    $componentsQuery.data
      ?.filter((c) => !c.component?.spec?.definedInCanvas)
      .map((c) => c.meta?.name?.name ?? "") ?? [];

  async function handleAddComponent() {
    const newFilePath = await createResourceFile(ResourceKind.Component);

    if (!newFilePath) return;

    const componentName = getNameFromFile(newFilePath);

    if (componentName) {
      addComponent(componentName);
    }
  }
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <Button builders={[builder]} type="secondary">
      <Plus class="flex items-center justify-center" size="16px" />
      <div class="flex gap-x-1 items-center">
        Add component
        <ChevronDown size="14px" />
      </div>
    </Button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content class="flex flex-col gap-y-1 ">
    <DropdownMenu.Group>
      <!-- <DropdownMenu.Item disabled>
        Generate chart
        {#if $ai}
          with AI
          <WandIcon class="w-3 h-3" />
        {/if}
      </DropdownMenu.Item> -->
      <DropdownMenu.Item on:click={handleAddComponent}>
        Create new component
      </DropdownMenu.Item>
    </DropdownMenu.Group>

    <DropdownMenu.Separator />
    <div class="px-1">
      <Search bind:value />
    </div>
    <DropdownMenu.Separator />

    <DropdownMenu.Label class="text-[11px] text-gray-500 py-0">
      EXISTING COMPONENTS
    </DropdownMenu.Label>
    <DropdownMenu.Group>
      {#each componentFileNames.filter( (n) => n.startsWith(value), ) as componentName (componentName)}
        <DropdownMenu.Item on:click={() => addComponent(componentName)}>
          {componentName}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>
