<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { PivotMeasureFormatting } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
  // When provided, measure chips expose per-measure conditional formatting
  // controls in a dropdown on the chip.
  export let measureFormatting:
    | Record<string, PivotMeasureFormatting>
    | undefined = undefined;
  export let setMeasureFormatting:
    | ((measureName: string, fmt: PivotMeasureFormatting | null) => void)
    | undefined = undefined;

  const client = useRuntimeClient();

  let open = false;
  let searchValue = "";

  $: ctx = getCanvasStore(canvasName, client.instanceId);
  $: fieldData = useMetricFieldData(ctx, metricName, types);

  $: metricsViewStore =
    ctx.canvasEntity.metricsView.getMetricsViewFromName(metricName);
  $: lowerIsBetterMap = Object.fromEntries(
    ($metricsViewStore.metricsView?.measures ?? []).map((m) => [
      m.name as string,
      m.lowerIsBetter ?? false,
    ]),
  );
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
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <div {...props} class="flex justify-between gap-x-2">
            <InputLabel small {label} {id} />
            <button
              aria-label={`Add ${types.join(", ")} fields`}
              class="text-sm px-2 h-6"
            >
              <PlusIcon size="14px" />
            </button>
          </div>
        {/snippet}
      </DropdownMenu.Trigger>
    </svelte:fragment>
  </FieldSelectorDropdown>

  <FieldChips
    items={selectedItems}
    displayMap={$fieldData.displayMap}
    onUpdate={onMultiSelect}
    {measureFormatting}
    {setMeasureFormatting}
    {lowerIsBetterMap}
  />
</div>
