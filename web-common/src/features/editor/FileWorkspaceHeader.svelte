<script lang="ts">
  import { goto } from "$app/navigation";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { PROTECTED_FILES } from "../file-explorer/protected-paths";

  export let filePath: string;
  const fileArtifact = fileArtifacts.getFileArtifact(filePath);

  let fileName: string;

  $: [, fileName] = splitFolderAndName(filePath);
  $: isProtectedFile = PROTECTED_FILES.includes(filePath);

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
      fileArtifacts.getNamesForKind(get(fileArtifact.name)?.kind), // TODO
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
