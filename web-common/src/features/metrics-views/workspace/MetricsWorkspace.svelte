<script lang="ts">
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appStore } from "@rilldata/web-common/layout/app-store";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import { WorkspaceContainer } from "../../../layout/workspace";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";
  import MetricsEditor from "./editor/MetricsEditor.svelte";
  import MetricsInspector from "./inspector/MetricsInspector.svelte";

  // the runtime yaml string
  export let metricsDefName: string;
  export let nonStandardError: string = undefined;

  // this store is used to store errors that are not related to the reconciliation/runtime
  // used to prevent the user from going to the dashboard.
  // Ultimately, the runtime should be catching the different errors we encounter with regards to
  // mismatches between the fields. For now, this is a very simple to use solution.
  let configurationErrorStore = writable({
    defaultTimeRange: null,
    smallestTimeGrain: null,
    model: null,
    timeColumn: null,
  });
  setContext("rill:metrics-config:errors", configurationErrorStore);

  const switchToMetrics = async (metricsDefName: string) => {
    if (!metricsDefName) return;

    appStore.setActiveEntity(metricsDefName, EntityType.MetricsDefinition);
  };

  $: switchToMetrics(metricsDefName);
</script>

<WorkspaceContainer inspector={true} assetID={`${metricsDefName}-config`}>
  <MetricsWorkspaceHeader slot="header" {metricsDefName} />
  <MetricsEditor slot="body" {metricsDefName} />
  <MetricsInspector slot="inspector" {metricsDefName} />
</WorkspaceContainer>
