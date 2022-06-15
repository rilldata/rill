<script lang="ts">
  import WorkspaceHeader from "../WorkspaceHeader.svelte";
  import { store, reduxReadable } from "$lib/redux-store/store-root";
  import { updateMetricDefLabel } from "$lib/redux-store/metrics-definition/metrics-definition-slice";

  export let metricsDefId;

  $: metricsDef = $reduxReadable?.metricsDefinition?.entities[metricsDefId];
  $: titleInput = metricsDef?.metricDefLabel;

  const onChangeCallback = async (e) => {
    store.dispatch(
      updateMetricDefLabel({ id: metricsDefId, label: e.target.value })
    );
  };
</script>

<WorkspaceHeader {...{ titleInput, onChangeCallback }} />
