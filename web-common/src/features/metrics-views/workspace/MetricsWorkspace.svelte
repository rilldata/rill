<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useIsModelingSupportedForCurrentOlapDriver } from "../../tables/selectors";
  import MetricsEditor from "./editor/MetricsEditor.svelte";
  import MetricsInspector from "./inspector/MetricsInspector.svelte";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";

  export let metricsDefName: string;

  $: isModelingSupportedForCurrentOlapDriver =
    useIsModelingSupportedForCurrentOlapDriver($runtime.instanceId);
  $: showInspector = $isModelingSupportedForCurrentOlapDriver.data;
</script>

<WorkspaceContainer inspector={showInspector}>
  <MetricsWorkspaceHeader
    slot="header"
    {metricsDefName}
    showInspectorToggle={showInspector}
  />
  <MetricsEditor slot="body" {metricsDefName} />
  <MetricsInspector slot="inspector" {metricsDefName} />
</WorkspaceContainer>
