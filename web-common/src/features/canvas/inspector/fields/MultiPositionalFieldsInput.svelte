<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { PlusIcon } from "lucide-svelte";
  import { useMetricFieldData } from "../selectors";
  import type { FieldType } from "../types";
  import FieldChips from "./FieldChips.svelte";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";

  export let canvasName: string;
  export let metricName: string;
  export let selectedItems: string[] = [];
  export let chipItems: string[] = [];
  export let types: FieldType[] = ["measure", "dimension"];
  export let excludedValues: string[] | undefined = undefined;
  export let onMultiSelect: (items: string[]) => void = () => {};

  let open = false;
  let searchValue = "";

  $: ({ instanceId } = $runtime);
  $: ctx = getCanvasStore(canvasName, instanceId);
  $: fieldData = useMetricFieldData(ctx, metricName, types);
</script>

<div class="w-full flex flex-col gap-y-2">
  <FieldChips
    items={chipItems}
    displayMap={$fieldData.displayMap}
    onUpdate={onMultiSelect}
  />

  <FieldSelectorDropdown
    {canvasName}
    {metricName}
    {selectedItems}
    {types}
    {excludedValues}
    {onMultiSelect}
    bind:open
    bind:searchValue
  >
    <svelte:fragment slot="trigger">
      <DropdownMenu.Trigger asChild let:builder>
        <div class="flex justify-between gap-x-2">
          <button
            aria-label={`Add ${types.join(", ")} fields`}
            use:builder.action
            {...builder}
            class="flex flex-row items-center gap-x-1 text-xs"
          >
            <PlusIcon size="14px" /> add measure
          </button>
        </div>
      </DropdownMenu.Trigger>
    </svelte:fragment>
  </FieldSelectorDropdown>
</div>
