<script lang="ts">
  import { goto } from "$app/navigation";
  import { splitFolderAndFileName } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/actions/ui-actions.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { WorkspaceHeader } from "../../layout/workspace";
  import type { ResourceKind } from "../entity-management/resource-selectors";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { readonlyFiles } from "@rilldata/web-common/features/entity-management/actions/readonly-files.ts";

  const runtimeClient = useRuntimeClient();

  let {
    filePath,
    hasUnsavedChanges,
    resourceKind,
    resource,
  }: {
    filePath: string;
    hasUnsavedChanges: boolean;
    resourceKind: ResourceKind | undefined;
    resource: V1Resource | undefined;
  } = $props();

  let [, fileName] = $derived(splitFolderAndFileName(filePath));
  let readonlyMatch = $derived(readonlyFiles.match(filePath));
  let isReadonly = $derived(!!readonlyMatch);

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
  editable={!isReadonly}
  nonEditableMessage={readonlyMatch?.messageSnippet}
  onTitleChange={onChangeCallback}
  {hasUnsavedChanges}
  showInspectorToggle={false}
  titleInput={fileName}
  {resource}
/>
