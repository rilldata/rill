<script lang="ts">
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { ChevronDown, Plus } from "lucide-svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { runtime } from "../../runtime-client/runtime-store";
  import { createEventDispatcher } from "svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";

  const dispatch = createEventDispatcher();
  let open = false;
  let value = "";

  // We want to get only valid charts here. Hence using ListResources API
  $: chartFileNamesQuery = useFilteredResources(
    $runtime.instanceId,
    ResourceKind.Chart,
    (data) => data.resources?.map((r) => r.meta?.name?.name ?? "") ?? [],
  );
  $: chartFileNames = $chartFileNamesQuery.data ?? [];
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <button {...builder} class:open use:builder.action>
      <Plus class="flex items-center justify-center" size="16px" />
      <div class="flex gap-x-1 items-center">
        Add Chart
        <ChevronDown size="14px" />
      </div>
    </button>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content class="flex flex-col gap-y-1 ">
    <DropdownMenu.Group>
      <DropdownMenu.Item disabled>Generate chart with AI</DropdownMenu.Item>
      <DropdownMenu.Item disabled>Specify new chart</DropdownMenu.Item>
    </DropdownMenu.Group>
    <DropdownMenu.Separator />
    <div class="px-1">
      <Search bind:value />
    </div>
    <DropdownMenu.Separator />
    <DropdownMenu.Label class="text-[11px] text-gray-500 py-0">
      EXISTING CHARTS
    </DropdownMenu.Label>
    <DropdownMenu.Group>
      {#each chartFileNames.filter( (n) => n.startsWith(value), ) as chartName (chartName)}
        <DropdownMenu.Item
          on:click={() => {
            dispatch("add-chart", { chartName });
          }}
        >
          {chartName}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  button {
    @apply w-fit h-8;
    @apply text-primary-600;
    @apply border-primary-300 border-2;
    @apply rounded-sm px-3 font-medium;
    @apply flex gap-x-2 items-center justify-center;
  }

  button:hover {
    @apply bg-primary-50;
  }

  button.open,
  button:active {
    @apply bg-primary-100;
  }
</style>
