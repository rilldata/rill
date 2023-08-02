<script lang="ts">
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { page } from "$app/stores";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import TimeDimensionDetailsTable from "./TimeDimensionDetailsTable.svelte";

  $: metricViewName = $page.params.name;
</script>

<WorkspaceContainer
  top="0px"
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <StateManagersProvider metricsViewName={metricViewName} slot="body">
    <div class="p-8">
      <TimeDimensionDetailsTable />
    </div>
  </StateManagersProvider>
</WorkspaceContainer>
