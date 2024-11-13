<script lang="ts">
  import { goto } from "$app/navigation";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { WorkspaceHeader } from "../../layout/workspace";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { PROTECTED_FILES } from "../file-explorer/protected-paths";

  export let filePath: string;
  export let hasUnsavedChanges: boolean;
  export let resourceKind: ResourceKind | undefined;

  let fileName: string;

  $: [, fileName] = splitFolderAndFileName(filePath);
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

  const onChangeCallback = async (newTitle: string) => {
    const route = await handleEntityRename(
      $runtime.instanceId,
      newTitle,
      filePath,
      fileName,
    );
    if (route) await goto(route);
  };
</script>

<WorkspaceHeader
  {filePath}
  {resourceKind}
  editable={!isProtectedFile}
  onTitleChange={onChangeCallback}
  {hasUnsavedChanges}
  showInspectorToggle={false}
  titleInput={fileName}
/>
