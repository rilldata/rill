<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    templates,
    type Template,
  } from "@rilldata/web-common/features/apis/editor/template-utils";
  import type { Arg } from "@rilldata/web-common/features/apis/editor/types";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    resourceIsLoading,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { mapParseErrorsToLines } from "@rilldata/web-common/features/metrics-views/errors";
  import APIEditor from "@rilldata/web-common/features/apis/editor/APIEditor.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceHeader from "@rilldata/web-common/layout/workspace/WorkspaceHeader.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ChevronDownIcon } from "lucide-svelte";

  export let fileArtifact: FileArtifact;

  const runtimeClient = useRuntimeClient();

  $: ({
    hasUnsavedChanges,
    autoSave,
    path: filePath,
    resourceName,
    remoteContent,
    fileName,
  } = fileArtifact);

  $: apiName = $resourceName?.name ?? getNameFromFile(filePath);
  $: host = runtimeClient.host || "http://localhost:9009";
  $: instanceId = runtimeClient.instanceId;

  $: allErrorsQuery = fileArtifact.getAllErrors(queryClient);
  $: allErrors = $allErrorsQuery;
  $: resourceQuery = fileArtifact.getResource(queryClient);
  $: ({ data: resource } = $resourceQuery);
  $: isReconciling = resourceIsLoading($resourceQuery.data);

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeClient,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }

  $: errors = mapParseErrorsToLines(allErrors, $remoteContent ?? "");

  let args: Arg[] = [];

  $: ({ updateEditorContent } = fileArtifact);

  // Template confirmation dialog state
  let dialogOpen = false;
  let pendingTemplate: Template | null = null;
  let dropdownOpen = false;

  function selectTemplate(template: Template) {
    pendingTemplate = template;
    dialogOpen = true;
  }

  function confirmTemplate() {
    if (pendingTemplate) {
      updateEditorContent(pendingTemplate.content);
    }
    dialogOpen = false;
    pendingTemplate = null;
  }
</script>

<WorkspaceContainer inspector={false}>
  <WorkspaceHeader
    {filePath}
    {resource}
    resourceKind={ResourceKind.API}
    hasUnsavedChanges={$hasUnsavedChanges}
    onTitleChange={onChangeCallback}
    slot="header"
    showInspectorToggle={false}
    titleInput={fileName}
  >
    <svelte:fragment slot="workspace-controls">
      <DropdownMenu.Root bind:open={dropdownOpen}>
        <DropdownMenu.Trigger asChild let:builder>
          <Tooltip distance={8} suppress={dropdownOpen}>
            <Button type="text" compact small builders={[builder]}>
              Templates
              <ChevronDownIcon size="12px" />
            </Button>
            <TooltipContent slot="tooltip-content">
              Insert a starter API template
            </TooltipContent>
          </Tooltip>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="end" class="w-56">
          {#each templates as template}
            <DropdownMenu.Item on:click={() => selectTemplate(template)}>
              <span class="text-sm">{template.label}</span>
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </svelte:fragment>
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    <APIEditor
      bind:autoSave={$autoSave}
      {fileArtifact}
      {errors}
      {apiName}
      {isReconciling}
      {host}
      {instanceId}
      bind:args
    />
  </svelte:fragment>
</WorkspaceContainer>

<AlertDialog bind:open={dialogOpen}>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{pendingTemplate?.label ?? "Template"}</AlertDialogTitle
      >
      <AlertDialogDescription>
        {pendingTemplate?.description ?? ""}
        {#if $hasUnsavedChanges}
          <br /><br />
          <strong>This will replace your current editor content.</strong>
        {/if}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          dialogOpen = false;
          pendingTemplate = null;
        }}
      >
        Cancel
      </Button>
      <Button type="primary" onClick={confirmTemplate}>Apply</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
