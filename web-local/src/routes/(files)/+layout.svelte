<script lang="ts">
  import FileExplorer from "@rilldata/web-common/layout/file-explorer/FileExplorer.svelte";
  export let data;
</script>

<main class="flex overflow-hidden h-screen w-screen absolute">
  <FileExplorer files={data.files} />

  <slot />
</main>
