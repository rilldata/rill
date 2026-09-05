<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PencilIcon, PlusIcon } from "lucide-svelte";
  import { useMetricFieldData } from "../selectors";
  import EphemeralMeasureEditor from "./EphemeralMeasureEditor.svelte";

  const client = useRuntimeClient();

  export let metricName: string;
  export let label: string | undefined = undefined;
  export let id: string;
  export let selectedItem: string | undefined = undefined;
  export let type: "measure" | "dimension";
  export let includeTime = false;
  export let canvasName: string;
  export let searchableItems: string[] | undefined = undefined;
  export let excludedValues: string[] | undefined = undefined;
  export let ephemeralMeasures: EphemeralMeasureDef[] | undefined = undefined;
  // Component owning the spec; when set (and this is a measure field), the
  // selector offers creating and editing ephemeral measures.
  export let component: BaseCanvasComponent | undefined = undefined;
  export let isRemovable = false;
  export let onSelect: (item: string, displayName: string) => void = () => {};
  export let onRemove: () => void = () => {};

  let open = false;
  let searchValue = "";
  let ephemeralEditorOpen = false;
  let ephemeralEditingDef: EphemeralMeasureDef | null = null;

  $: ephemeralEnabled = !!component && type === "measure";

  function openEphemeralEditor(def: EphemeralMeasureDef | null) {
    ephemeralEditingDef = def;
    ephemeralEditorOpen = true;
    open = false;
  }

  $: ephemeralDefsByName = new Map(
    (ephemeralMeasures ?? []).map((def) => [def.name, def]),
  );

  $: ctx = getCanvasStore(canvasName, client.instanceId);
  $: ({ getTimeDimensionForMetricView } = ctx.canvasEntity.metricsView);

  $: timeDimension = getTimeDimensionForMetricView(metricName);

  $: isTimeSelected = $timeDimension && selectedItem === $timeDimension;
  $: effectiveExcludedValues =
    type === "dimension" && !includeTime && $timeDimension
      ? [...(excludedValues ?? []), $timeDimension]
      : excludedValues;
  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    [type],
    searchableItems,
    searchValue,
    effectiveExcludedValues,
    type === "measure" ? ephemeralMeasures : undefined,
  );
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <div class="flex items-center gap-x-2">
    {#if label}
      <InputLabel small {label} {id} />
    {/if}
  </div>

  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Chip
          {...props}
          fullWidth
          caret
          removable={isRemovable && !!selectedItem}
          {onRemove}
          type={isTimeSelected ? "time" : type}
        >
          <span
            class="font-bold truncate inline-flex items-center gap-x-1"
            class:text-fg-tertiary={!selectedItem}
            slot="body"
          >
            {#if isTimeSelected}
              {m.canvas_field_time()}
            {:else if selectedItem}
              {$fieldData.displayMap[selectedItem]?.label || selectedItem}
              {#if ephemeralDefsByName.has(selectedItem)}
                <span class="flex-none text-[10px] font-semibold italic">
                  ƒx
                </span>
              {/if}
            {:else}
              {m.canvas_field_choose_field()}
            {/if}
          </span>
        </Chip>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content class="p-0" sameWidth>
      <div class="p-3 pb-1">
        <Search bind:value={searchValue} autofocus={false} />
      </div>
      <div class="max-h-64 overflow-y-auto pb-2">
        {#if type == "dimension" && includeTime && $timeDimension}
          <DropdownMenu.Item
            class="pl-8 mx-1"
            onclick={() => {
              onSelect($timeDimension, m.canvas_field_time());
              open = false;
            }}
          >
            {m.canvas_field_time()}
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
        {/if}
        {#each $fieldData.filteredItems as item (item)}
          {#if item !== selectedItem}
            {@const ephemeralDef = ephemeralDefsByName.get(item)}
            <DropdownMenu.Item
              class="pl-8 mx-1"
              onclick={() => {
                onSelect(item, $fieldData.displayMap[item]?.label || item);
                open = false;
              }}
            >
              <slot {item}>
                <span
                  class="inline-flex w-full items-center gap-x-1"
                  title={ephemeralDef?.expression}
                >
                  {$fieldData.displayMap[item]?.label || item}
                  {#if ephemeralDef}
                    <span class="text-[10px] font-semibold italic">ƒx</span>
                    {#if ephemeralEnabled}
                      <button
                        class="ml-auto flex-none text-fg-secondary hover:text-fg-primary"
                        type="button"
                        aria-label={m.dashboard_pivot_ephemeral_edit_title()}
                        onpointerdown={(e) => e.stopPropagation()}
                        onpointerup={(e) => e.stopPropagation()}
                        onclick={(e) => {
                          e.stopPropagation();
                          openEphemeralEditor(ephemeralDef);
                        }}
                      >
                        <PencilIcon size="12px" />
                      </button>
                    {/if}
                  {/if}
                </span>
              </slot>
            </DropdownMenu.Item>
          {/if}
        {:else}
          {#if searchValue}
            <div class="text-fg-disabled text-center p-2 w-full">
              {m.canvas_field_no_results()}
            </div>
          {/if}
        {/each}
      </div>
      {#if ephemeralEnabled}
        <footer class="border-t bg-gray-100">
          <button
            class="flex w-full items-center gap-x-1.5 h-8 px-3 text-xs text-fg-primary hover:bg-surface-hover"
            type="button"
            onclick={() => openEphemeralEditor(null)}
          >
            <PlusIcon size="14px" />
            {m.dashboard_pivot_ephemeral_create()}
          </button>
        </footer>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>

{#if ephemeralEditorOpen && component}
  <EphemeralMeasureEditor
    {component}
    {canvasName}
    {metricName}
    editingDef={ephemeralEditingDef}
    onClose={() => (ephemeralEditorOpen = false)}
    onCreated={(name, displayName) => onSelect(name, displayName || name)}
  />
{/if}
