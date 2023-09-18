<script>
  import YamlWorkspace from "@rilldata/web-common/features/generic-yaml-editor/YamlWorkspace.svelte";

  const fileName = "rill.yaml";
</script>

<svelte:head>
  <title>{fileName}</title>
</svelte:head>

<YamlWorkspace {fileName} />
