<script lang="ts">
  import {
    ResourceKind,
    useClientFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { ChevronDown, Plus } from "lucide-svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { runtime } from "../../runtime-client/runtime-store";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { getNameFromFile } from "../entity-management/entity-mappers";
  // import { featureFlags } from "../feature-flags";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { handleEntityCreate } from "../file-explorer/new-files";

  // const { ai } = featureFlags;

  export let addChart: (chartName: string) => void;

  let open = false;
  let value = "";

  // We want to get only valid charts here. Hence using ListResources API
  $: chartsQuery = useClientFilteredResources(
    $runtime.instanceId,
    ResourceKind.Component,
  );
  $: chartFileNames =
    $chartsQuery.data?.map((c) => c.meta?.name?.name ?? "") ?? [];

  async function handleAddChart() {
    const newRoute = await handleEntityCreate(ResourceKind.Component);

    if (!newRoute) return;

    const chartName = getNameFromFile(newRoute);

    if (chartName) {
      addChart(chartName);
    }
  }
</script>

<DropdownMenu.Root bind:open typeahead={false}>
  <DropdownMenu.Trigger asChild let:builder>
    <Button builders={[builder]} type="secondary">
      <Plus class="flex items-center justify-center" size="16px" />
      <div class="flex gap-x-1 items-center">
        Add chart
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
      <DropdownMenu.Item on:click={handleAddChart}>
        Create new chart
      </DropdownMenu.Item>
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
        <DropdownMenu.Item on:click={() => addChart(chartName)}>
          {chartName}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>
