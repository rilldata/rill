<script lang="ts">
  import { deleteFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import RenameAssetModal from "@rilldata/web-common/features/entity-management/RenameAssetModal.svelte";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
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
  $: instanceId = $runtime.instanceId;

  let showRenameModelModal = false;
  let renameFilePath: string;
  function onRename(filePath: string) {
    showRenameModelModal = true;
    renameFilePath = filePath;
  }

  async function onDelete(filePath: string) {
    await deleteFileArtifact(instanceId, filePath, true);
  }
</script>

<div class="flex flex-col items-start gap-y-2">
  <!-- File tree -->
  <div class="flex flex-col w-full items-start justify-start overflow-auto">
    {#if $getFileTree.data}
      <NavDirectory directory={$getFileTree.data} {onRename} {onDelete} />
    {/if}
  </div>
</div>

{#if showRenameModelModal}
  <RenameAssetModal
    closeModal={() => (showRenameModelModal = false)}
    filePath={renameFilePath}
  />
{/if}
