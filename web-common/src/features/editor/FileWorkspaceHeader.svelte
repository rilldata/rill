<script lang="ts">
  import { goto } from "$app/navigation";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  import { WorkspaceHeader } from "../../layout/workspace";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { PROTECTED_FILES } from "../file-explorer/protected-paths";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let filePath: string;
  export let hasUnsavedChanges: boolean;
  export let resourceKind: ResourceKind | undefined;
  export let resource: V1Resource | undefined;

  let fileName: string;

  $: [, fileName] = splitFolderAndFileName(filePath);
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

  $: ({ instanceId } = runtimeClient);

  const onChangeCallback = async (newTitle: string) => {
    const route = await handleEntityRename(
      instanceId,
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
  {resource}
/>
