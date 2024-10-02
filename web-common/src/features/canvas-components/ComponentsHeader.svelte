<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import GenerateVegaSpecPrompt from "@rilldata/web-common/features/canvas-components/prompt/GenerateVegaSpecPrompt.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    extractFileName,
    splitFolderAndName,
  } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let filePath: string;
  export let hasUnsavedChanges: boolean;

  let fileName: string;
  $: [, fileName] = splitFolderAndName(filePath);
  $: runtimeInstanceId = $runtime.instanceId;
  $: componentName = extractFileName(filePath);

  async function handleNameChange(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const newRoute = await handleEntityRename(
      runtimeInstanceId,
      e.currentTarget,
      filePath,
      componentName,
      fileArtifacts.getNamesForKind(ResourceKind.Component),
    );

    if (newRoute) await goto(newRoute);
  }

  let generateOpen = false;
</script>

<WorkspaceHeader
  resourceKind={ResourceKind.Component}
  on:change={handleNameChange}
  titleInput={fileName}
  {hasUnsavedChanges}
  {filePath}
  showInspectorToggle={false}
>
  <svelte:fragment slot="cta">
    <PanelCTA side="right">
      <Button on:click={() => (generateOpen = true)}>Generate using AI</Button>
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>

<GenerateVegaSpecPrompt
  bind:open={generateOpen}
  chart={componentName}
  {filePath}
/>
