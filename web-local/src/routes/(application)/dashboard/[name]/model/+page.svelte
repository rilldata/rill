<script lang="ts">
  import { page } from "$app/stores";
  import ModelBody from "@rilldata/web-common/features/models/workspace/ModelBody.svelte";
  import WorkspaceBody from "@rilldata/web-local/lib/components/workspace/core/WorkspaceBody.svelte";
  export let data;
  $: metricViewName = $page?.params?.name;
  $: modelName = data?.entry?.metricsView?.model;
</script>

<svelte:head>
  <title>Rill Developer | {modelName}</title>
</svelte:head>
<WorkspaceBody>
  <ModelBody {modelName} />
</WorkspaceBody>
<!-- <ModelWorkspace {modelName} focusEditorOnMount={true} /> -->
