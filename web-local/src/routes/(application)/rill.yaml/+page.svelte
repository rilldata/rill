<script>
  import YamlWorkspace from "@rilldata/web-common/components/editor/workspace/YamlWorkspace.svelte";

  const fileName = "rill.yaml";
</script>

<svelte:head>
  <title>{fileName}</title>
</svelte:head>

<YamlWorkspace {fileName} />
