<script lang="ts">
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import NavDirectory from "./NavDirectory.svelte";
  import { transformFileList } from "./transform-file-list";

  $: getFileTree = createRuntimeServiceListFiles("default", undefined, {
    query: {
      select: (data) => {
        if (!data || !data.files?.length) return;

        const files = data.files
          // remove leading slash
          .map((file) => ({
            path: file.path?.slice(1) ?? "",
            isDir: !!file.isDir,
          }))
          // sort alphabetically case-insensitive
          .sort((a, b) =>
            a.path.localeCompare(b.path, undefined, { sensitivity: "base" }),
          );

        const fileTree = transformFileList(files);

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
