<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import {
    extractFileName,
    splitFolderAndFileName,
  } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import GenerateVegaSpecPrompt from "./prompt/GenerateVegaSpecPrompt.svelte";

  export let filePath: string;
  export let hasUnsavedChanges: boolean;

  let fileName: string;
  $: [, fileName] = splitFolderAndFileName(filePath);
  $: runtimeInstanceId = $runtime.instanceId;
  $: componentName = extractFileName(filePath);

  async function handleNameChange(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeInstanceId,
      newTitle,
      filePath,
      componentName,
    );

    if (newRoute) await goto(newRoute);
  }

  let generateOpen = false;
</script>

<WorkspaceHeader
  resourceKind={ResourceKind.Component}
  onTitleChange={handleNameChange}
  titleInput={fileName}
  {hasUnsavedChanges}
  {filePath}
  showInspectorToggle={false}
>
  <svelte:fragment slot="cta">
    <PanelCTA side="right">
      <Button type="secondary" on:click={() => (generateOpen = true)}
        >Generate using AI</Button
      >
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>

<GenerateVegaSpecPrompt
  bind:open={generateOpen}
  chart={componentName}
  {filePath}
/>
