<script lang="ts">
  import { goto } from "$app/navigation";
  import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { WorkspaceHeader } from "../../layout/workspace";

  export let filePath: string;
  let fileName: string;
  $: [, fileName] = splitFolderAndName(filePath);

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
    );
    if (route) await goto(route);
  };
</script>

<WorkspaceHeader
  on:change={onChangeCallback}
  showInspectorToggle={false}
  titleInput={fileName}
/>
