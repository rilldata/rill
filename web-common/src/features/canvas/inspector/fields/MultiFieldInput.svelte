<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { PlusIcon } from "lucide-svelte";
  import { useMetricFieldData } from "../selectors";
  import type { FieldType } from "../types";
  import FieldChips from "./FieldChips.svelte";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";

  export let canvasName: string;
  export let metricName: string;
  export let label: string;
  export let id: string;
  export let selectedItems: string[] = [];
  export let types: FieldType[];
  export let onMultiSelect: (items: string[]) => void = () => {};

  let open = false;
  let searchValue = "";

  $: ({ instanceId } = $runtime);
  $: ctx = getCanvasStore(canvasName, instanceId);
  $: fieldData = useMetricFieldData(ctx, metricName, types);
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <FieldSelectorDropdown
    {canvasName}
    {metricName}
    {selectedItems}
    {types}
    {onMultiSelect}
    bind:open
    bind:searchValue
  >
    <svelte:fragment slot="trigger">
      <DropdownMenu.Trigger asChild let:builder>
        <div class="flex justify-between gap-x-2">
          <InputLabel small {label} {id} />
          <button
            aria-label={`Add ${types.join(", ")} fields`}
            use:builder.action
            {...builder}
            class="text-sm px-2 h-6"
          >
            <PlusIcon size="14px" />
          </button>
        </div>
      </DropdownMenu.Trigger>
    </svelte:fragment>
  </FieldSelectorDropdown>

  <FieldChips
    items={selectedItems}
    displayMap={$fieldData.displayMap}
    onUpdate={onMultiSelect}
  />
</div>
