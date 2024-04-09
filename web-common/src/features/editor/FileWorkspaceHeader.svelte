<script lang="ts">
  import { goto } from "$app/navigation";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { EntityType } from "../entity-management/types";

  export let filePath: string;
  $: entityName = extractFileName(filePath);

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    const route = await handleEntityRename(
      $runtime.instanceId,
      e.currentTarget,
      filePath,
      EntityType.Unknown,
    );
    if (route) await goto(route);
  };
</script>

<WorkspaceHeader
  on:change={onChangeCallback}
  showInspectorToggle={false}
  titleInput={entityName}
/>
