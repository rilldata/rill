<script lang="ts">
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import NavDirectory from "./NavDirectory.svelte";
  import { transformFileList } from "./transform-file-list";

  $: getFileTree = createRuntimeServiceListFiles("default", undefined, {
    query: {
      select: (data) => {
        if (!data || !data.paths) return;

        const paths = data.paths
          // remove leading slash
          .map((path) => path.slice(1))
          // sort alphabetically case-insensitive
          .sort((a, b) =>
            a.localeCompare(b, undefined, { sensitivity: "base" }),
          );

        const fileTree = transformFileList(paths);

        return fileTree;
      },
    },
  });
</script>

<div class="flex flex-col items-start gap-y-2">
  <!-- File tree -->
  <div class="flex flex-col w-full items-start justify-start overflow-auto">
    {#if $getFileTree.data}
      <NavDirectory directory={$getFileTree.data} />
    {/if}
  </div>
</div>
