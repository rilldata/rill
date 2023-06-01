<script lang="ts">
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import NavDirectory from "./NavDirectory.svelte";
  import { transformFileList } from "./transform-file-list";

  $: getFileTree = createRuntimeServiceListFiles("default", undefined, {
    query: {
      select: (data) => {
        const paths = data.paths
          // remove leading slash
          .map((path) => path.slice(1))
          // sort alphabetically case-insensitive
          .sort((a, b) =>
            a.localeCompare(b, undefined, { sensitivity: "base" })
          );
        const fileTree = transformFileList(paths);
        return fileTree;
      },
    },
  });

  // Button handler: Create new file
  // const createFile = createRuntimeServicePutFileAndReconcile();
  // let path: string;
  // function submit() {
  //   $createFile.mutateAsync({
  //     data: {
  //       instanceId: "default",
  //       path: path,
  //       blob: undefined,
  //       create: true,
  //       createOnly: true,
  //       strict: false,
  //     },
  //   });
  // }
</script>

<div class="flex flex-col items-start gap-y-2">
  <!-- <div class="flex gap-x-2">
    <input type="text" bind:value={path} class="border border-blue-500" />
    <button on:click={submit}>Create new file</button>
  </div> -->

  <!-- File tree -->
  <div
    class="
  flex
  flex-col
  w-full
  items-start
  justify-start
  overflow-auto
  "
  >
    {#if $getFileTree.data}
      <NavDirectory directory={$getFileTree.data} />
    {/if}
  </div>
</div>
