<script lang="ts">
  import { ModelWorkspace } from "@rilldata/web-common/features/models";

  export let data;
</script>

<svelte:head>
  <title>Rill Developer | {data.modelName}</title>
</svelte:head>

<ModelWorkspace
  modelName={data.modelName}
  focusEditorOnMount={data.focusEditor}
/>
