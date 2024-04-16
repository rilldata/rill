<script>
  import YamlWorkspace from "@rilldata/web-common/features/generic-yaml-editor/YamlWorkspace.svelte";

  const filePath = "/rill.yaml";
</script>

<svelte:head>
  <title>{filePath}</title>
</svelte:head>

<YamlWorkspace {filePath} />
