<script lang="ts">
  import { goto } from "$app/navigation";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { WorkspaceHeader } from "../../layout/workspace";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import { PROTECTED_FILES } from "../file-explorer/protected-paths";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import type { Snippet } from "svelte";

  const runtimeClient = useRuntimeClient();

  let {
    filePath,
    hasUnsavedChanges,
    resourceKind,
    resource,
    editable = true,
    nonEditableMessage,
  }: {
    filePath: string;
    hasUnsavedChanges: boolean;
    resourceKind: ResourceKind | undefined;
    resource: V1Resource | undefined;
    editable: boolean;
    nonEditableMessage?: Snippet;
  } = $props();

  let [, fileName] = $derived(splitFolderAndFileName(filePath));
  let isProtectedFile = $derived(PROTECTED_FILES.includes(filePath));

  const onChangeCallback = async (newTitle: string) => {
    const route = await handleEntityRename(
      runtimeClient,
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
  editable={!isProtectedFile && editable}
  nonEditableMessage={!editable ? nonEditableMessage : undefined}
  onTitleChange={onChangeCallback}
  {hasUnsavedChanges}
  showInspectorToggle={false}
  titleInput={fileName}
  {resource}
/>
