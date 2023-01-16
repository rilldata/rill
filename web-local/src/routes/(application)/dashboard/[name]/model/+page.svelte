<script lang="ts">
  import { page } from "$app/stores";
  import { ModelWorkspace } from "@rilldata/web-common/features/models";

  export let data;

  $: metricViewName = $page?.params?.name;
  $: modelName = data?.entry?.metricsView?.model;
</script>

<svelte:head>
  <title>Rill Developer | {modelName}</title>
</svelte:head>

{modelName}
{metricViewName}

<ModelWorkspace {modelName} focusEditorOnMount={true} />
