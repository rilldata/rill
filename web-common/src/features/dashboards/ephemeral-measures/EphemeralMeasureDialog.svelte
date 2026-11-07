<!-- @component
Creates or edits an ephemeral measure on the explore dashboard's state.
-->
<script lang="ts">
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { get } from "svelte/store";
  import { ephemeralMeasureDialog } from "./dialog-store";
  import EphemeralMeasureForm from "./EphemeralMeasureForm.svelte";
  import type { EphemeralMeasureDef } from "./types";
  import { isReferenceableMeasure } from "./validation";

  // Mounted only while the dialog is open (see PivotDisplay), so the form
  // state initializes fresh from the store on every open.
  const editingDef = get(ephemeralMeasureDialog)?.def ?? null;

  const { exploreName, validSpecStore, dashboardStore } = getStateManagers();

  $: metricsView = $validSpecStore.data?.metricsView;
  $: explore = $validSpecStore.data?.explore;

  // Measures the expression may reference: the metrics view's simple measures
  // that are part of this explore (window / time-comparison / required-dimension
  // measures are excluded; see isReferenceableMeasure).
  $: referenceableMeasures = (metricsView?.measures ?? []).filter(
    (mes) =>
      mes.name &&
      (explore?.measures ?? []).includes(mes.name) &&
      isReferenceableMeasure(mes),
  );

  function close() {
    ephemeralMeasureDialog.set(null);
  }

  function save(def: EphemeralMeasureDef) {
    if (editingDef) {
      metricsExplorerStore.updateEphemeralMeasure($exploreName, def);
    } else {
      metricsExplorerStore.addEphemeralMeasure($exploreName, def);
    }
    close();
  }

  function remove() {
    if (!editingDef) return;
    metricsExplorerStore.removeEphemeralMeasure(
      $exploreName,
      editingDef.name,
      explore,
    );
    close();
  }
</script>

<EphemeralMeasureForm
  {editingDef}
  {metricsView}
  {referenceableMeasures}
  existingDefs={$dashboardStore?.ephemeralMeasures ?? []}
  onSave={save}
  onRemove={editingDef ? remove : undefined}
  onClose={close}
/>
