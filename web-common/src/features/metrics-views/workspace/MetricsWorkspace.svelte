<script lang="ts">
  import { WorkspaceContainer } from "../../../layout/workspace";
  import MetricsEditor from "./editor/MetricsEditor.svelte";
  import MetricsInspector from "./inspector/MetricsInspector.svelte";
  import MetricsWorkspaceHeader from "./MetricsWorkspaceHeader.svelte";

  export let metricsDefName: string;
</script>

<WorkspaceContainer>
  <MetricsWorkspaceHeader slot="header" {metricsDefName} />
  <MetricsEditor slot="body" {metricsDefName} />
  <MetricsInspector slot="inspector" {metricsDefName} />
</WorkspaceContainer>
