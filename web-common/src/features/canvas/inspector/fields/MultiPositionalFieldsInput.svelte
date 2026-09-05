<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import type { EphemeralMeasureDef } from "@rilldata/web-common/features/dashboards/ephemeral-measures/types";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { PlusIcon } from "lucide-svelte";
  import { useMetricFieldData } from "../selectors";
  import type { FieldType } from "../types";
  import EphemeralMeasureEditor from "./EphemeralMeasureEditor.svelte";
  import FieldChips from "./FieldChips.svelte";
  import FieldSelectorDropdown from "./FieldSelectorDropdown.svelte";

  export let canvasName: string;
  export let metricName: string;
  export let selectedItems: string[] = [];
  export let chipItems: string[] = [];
  export let types: FieldType[] = ["measure", "dimension"];
  export let excludedValues: string[] | undefined = undefined;
  export let ephemeralMeasures: EphemeralMeasureDef[] | undefined = undefined;
  // Component owning the spec; when set (and the field accepts measures), the
  // selector offers creating and editing ephemeral measures.
  export let component: BaseCanvasComponent | undefined = undefined;
  export let onMultiSelect: (items: string[]) => void = () => {};

  const client = useRuntimeClient();

  let open = false;
  let searchValue = "";
  let ephemeralEditorOpen = false;
  let ephemeralEditingDef: EphemeralMeasureDef | null = null;

  $: ephemeralEnabled = !!component && types.includes("measure");

  function openEphemeralEditor(def: EphemeralMeasureDef | null) {
    ephemeralEditingDef = def;
    ephemeralEditorOpen = true;
  }

  $: ctx = getCanvasStore(canvasName, client.instanceId);
  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    types,
    undefined,
    "",
    undefined,
    ephemeralMeasures,
  );
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
    {ephemeralMeasures}
    {onMultiSelect}
    onCreateEphemeral={ephemeralEnabled
      ? () => openEphemeralEditor(null)
      : undefined}
    onEditEphemeral={ephemeralEnabled
      ? (name) => {
          const def = ephemeralMeasures?.find((d) => d.name === name);
          if (def) openEphemeralEditor(def);
        }
      : undefined}
    bind:open
    bind:searchValue
  >
    <svelte:fragment slot="trigger">
      <DropdownMenu.Trigger>
        {#snippet child({ props })}
          <div {...props} class="flex justify-between gap-x-2">
            <button
              aria-label={`Add ${types.join(", ")} fields`}
              class="flex flex-row items-center gap-x-1 text-xs"
            >
              <PlusIcon size="14px" /> add measure
            </button>
          </div>
        {/snippet}
      </DropdownMenu.Trigger>
    </svelte:fragment>
  </FieldSelectorDropdown>
</div>

{#if ephemeralEditorOpen && component}
  <EphemeralMeasureEditor
    {component}
    {canvasName}
    {metricName}
    editingDef={ephemeralEditingDef}
    onClose={() => (ephemeralEditorOpen = false)}
    onCreated={(name) => onMultiSelect([...selectedItems, name])}
  />
{/if}
