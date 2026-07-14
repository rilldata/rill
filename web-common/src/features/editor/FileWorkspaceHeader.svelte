<script lang="ts">
  import { goto } from "$app/navigation";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/actions/ui-actions.ts";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { WorkspaceHeader } from "../../layout/workspace";
  import type { ResourceKind } from "../entity-management/resource-selectors";

  const runtimeClient = useRuntimeClient();

  let {
    fileArtifact,
    hasUnsavedChanges,
    resourceKind,
    resource,
  }: {
    fileArtifact: FileArtifact;
    hasUnsavedChanges: boolean;
    resourceKind: ResourceKind | undefined;
    resource: V1Resource | undefined;
  } = $props();

  let filePath = $derived(fileArtifact.path);
  let [, fileName] = $derived(splitFolderAndFileName(filePath));
  let editable = $derived(!fileArtifact.managed && !fileArtifact.pinned);

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
  {editable}
  onTitleChange={onChangeCallback}
  {hasUnsavedChanges}
  showInspectorToggle={false}
  titleInput={fileName}
  {resource}
/>
