<script lang="ts" context="module">
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { getStateManagers } from "../state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import type { PivotChipData } from "./types";
</script>

<script lang="ts">
  export let zone: "rows" | "columns" | null = null;

  const {
    selectors: {
      pivot: { dimensions, measures },
    },
    metricsViewName,
  } = getStateManagers();

  let open = false;

  function handleSelectValue(data: PivotChipData) {
    metricsExplorerStore.addPivotField($metricsViewName, data, zone === "rows");
  }
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger asChild let:builder>
    <Button builders={[builder]} type="add" selected={open}>
      <Add size="17px" />
    </Button>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content class="h-80 w-64 overflow-scroll" align="start">
    {#if zone === "columns"}
      <DropdownMenu.Label>Measures</DropdownMenu.Label>
      <DropdownMenu.Group>
        {#each $measures as measure}
          <DropdownMenu.Item
            on:click={() => {
              handleSelectValue(measure);
            }}
          >
            {measure.title}
          </DropdownMenu.Item>
        {/each}
      </DropdownMenu.Group>
      <DropdownMenu.Separator />
    {/if}
    <DropdownMenu.Label>Dimensions</DropdownMenu.Label>
    <DropdownMenu.Group>
      {#each $dimensions as dimension}
        <DropdownMenu.Item
          on:click={() => {
            handleSelectValue(dimension);
          }}
        >
          {dimension.title}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>
