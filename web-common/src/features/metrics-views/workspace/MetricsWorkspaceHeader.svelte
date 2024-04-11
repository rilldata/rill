<script lang="ts">
  import { goto } from "$app/navigation";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import GoToDashboardButton from "./GoToDashboardButton.svelte";

  export let filePath: string;
  export let showInspectorToggle = true;

  $: metricsDefName = extractFileName(filePath);

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    const newRoute = await handleEntityRename(
      $runtime.instanceId,
      e.currentTarget,
      filePath,
      metricsDefName,
    );
    if (newRoute) await goto(newRoute);
  };
</script>

<WorkspaceHeader
  on:change={onChangeCallback}
  {showInspectorToggle}
  titleInput={metricsDefName}
>
  <GoToDashboardButton {filePath} slot="cta" />
</WorkspaceHeader>
