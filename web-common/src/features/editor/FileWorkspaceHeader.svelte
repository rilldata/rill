<script lang="ts">
  import { goto } from "$app/navigation";
  import { useFileNamesInDirectory } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { PROTECTED_FILES } from "../file-explorer/protected-paths";

  export let filePath: string;

  let fileName: string;
  let folder: string;

  $: [folder, fileName] = splitFolderAndName(filePath);
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

  $: currentDirectoryFileNamesQuery = useFileNamesInDirectory(
    $runtime.instanceId,
    folder,
  );

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    const route = await handleEntityRename(
      $runtime.instanceId,
      e.currentTarget,
      filePath,
      fileName,
      $currentDirectoryFileNamesQuery.data ?? [],
    );
    if (route) await goto(route);
  };
</script>

<WorkspaceHeader
  editable={!isProtectedFile}
  on:change={onChangeCallback}
  showInspectorToggle={false}
  titleInput={fileName}
/>
