<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import FileWorkspaceHeader from "@rilldata/web-common/features/editor/FileWorkspaceHeader.svelte";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import { getExtensionsForFile } from "@rilldata/web-common/features/editor/getExtensionsForFile";
  import EnvVariablesTable from "./EnvVariablesTable.svelte";
  import type { EnvVariable } from "./types";
  import { Code2Icon, Plus, Settings, Download, Upload } from "lucide-svelte";
  import { parse as parseDotenv } from "dotenv";
  import type { EditorView } from "@codemirror/view";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { flexRender } from "tanstack-table-8-svelte-5";
  import ActionsCell from "./ActionsCell.svelte";
  import AddEnvDialog from "./AddEnvDialog.svelte";
  import PullEnvDialog from "./PullEnvDialog.svelte";
  import PushEnvDialog from "./PushEnvDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGetMetadata,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { getManageProjectAccess } from "@rilldata/web-common/features/project/selectors";
  import { getCloudFrontendUrl } from "@rilldata/web-common/features/organization/utils";

  export let fileArtifact: any;

  let editor: EditorView;
  let viewMode: "code" | "viz" = "viz";
  let addDialogOpen = false;
  let pullDialogOpen = false;
  let pushDialogOpen = false;

  const currentProjectQuery = createLocalServiceGetCurrentProject();
  const metadataQuery = createLocalServiceGetMetadata();

  $: project = $currentProjectQuery.data?.project;
  $: isProjectLinked = Boolean(project?.orgName && project?.name);
  $: manageProjectAccess = project
    ? getManageProjectAccess(project.orgName ?? "", project.name ?? "")
    : null;

  $: rcUrl = (() => {
    const currentProject = $currentProjectQuery.data;
    const metadata = $metadataQuery.data;
    const project = currentProject?.project;
    const adminUrl = metadata?.adminUrl;
    const hasManageProject = $manageProjectAccess ?? false;

    // Only show RC URL if user has manageProject permission (is admin)
    if (!project?.orgName || !project?.name || !adminUrl || !hasManageProject)
      return "";

    const url = new URL(getCloudFrontendUrl(adminUrl));
    url.pathname = `/${project.orgName}/${project.name}/-/settings/environment-variables`;
    return url.toString();
  })();

  $: ({ autoSave, hasUnsavedChanges, path, remoteContent, editorContent } =
    fileArtifact);

  $: extensions = getExtensionsForFile(path);

  // Use editorContent if it exists (user has made edits), otherwise use remoteContent
  $: currentContent = $editorContent || $remoteContent;
  $: envVariables = currentContent ? parseEnvFile(currentContent) : [];

  function parseEnvFile(content: string): EnvVariable[] {
    try {
      const parsed = parseDotenv(content);
      return Object.entries(parsed).map(([key, value]) => ({
        key,
        value: value ?? "",
      }));
    } catch (error) {
      console.error("Error parsing .env file:", error);
      return [];
    }
  }

  function serializeEnvFile(variables: EnvVariable[]): string {
    return variables.map((v) => `${v.key}=${v.value}`).join("\n") + "\n";
  }

  async function updateEnvFile(variables: EnvVariable[]) {
    const newContent = serializeEnvFile(variables);
    // Update editor content without autosave
    fileArtifact.updateEditorContent(newContent, false, false);
    // Force save since .env has autosave disabled
    await fileArtifact.saveLocalContent(true);
  }

  function handleToggleView() {
    viewMode = viewMode === "code" ? "viz" : "code";
  }

  async function handleAddVariables(
    event: CustomEvent<{ variables: EnvVariable[] }>,
  ) {
    const newVariables = event.detail.variables;
    const updatedVariables = [...envVariables, ...newVariables];
    await updateEnvFile(updatedVariables);
    eventBus.emit("notification", {
      type: "success",
      message: `Added ${newVariables.length} variable${newVariables.length === 1 ? "" : "s"}.`,
    });
  }

  async function handleEditVariable(
    oldKey: string,
    key: string,
    value: string,
  ) {
    const updatedVariables = envVariables.map((v) =>
      v.key === oldKey ? { key, value } : v,
    );
    await updateEnvFile(updatedVariables);
    eventBus.emit("notification", {
      type: "success",
      message: `Updated variable "${key}".`,
    });
  }

  async function handleDeleteVariable(key: string) {
    const updatedVariables = envVariables.filter((v) => v.key !== key);
    await updateEnvFile(updatedVariables);
    eventBus.emit("notification", {
      type: "success",
      message: `Deleted variable "${key}".`,
    });
  }

  $: actionsColumn = {
    accessorKey: "actions",
    header: "",
    cell: ({ row }: any) => {
      return flexRender(ActionsCell, {
        keyName: row.original.key,
        value: row.original.value,
        existingVariables: envVariables,
        onSave: handleEditVariable,
        onDelete: handleDeleteVariable,
      });
    },
    enableSorting: false,
  } as ColumnDef<EnvVariable, any>;
</script>

<WorkspaceContainer inspector={false}>
  <FileWorkspaceHeader
    slot="header"
    resource={undefined}
    resourceKind={undefined}
    filePath={path}
    hasUnsavedChanges={$hasUnsavedChanges}
    showIcon={false}
  >
    <div slot="left" class="radio relative mr-1">
      {#each [{ view: "code", icon: Code2Icon, label: "Code view" }, { view: "viz", icon: Settings, label: "No-code view" }] as { view, icon: Icon, label } (view)}
        <Tooltip activeDelay={700} distance={8}>
          <button
            aria-label="Switch to {label}"
            id="{view}-toggle"
            class="size-[22px] z-10 hover:brightness-75 p-0"
            on:click={handleToggleView}
          >
            <Icon size="15px" />
          </button>
          <TooltipContent slot="tooltip-content">
            {label}
          </TooltipContent>
        </Tooltip>
      {/each}
      <span
        style:left={viewMode === "code" ? "2px" : "24px"}
        class="toggle size-[22px] pointer-events-none absolute rounded-[4px] z-0 transition-[left]"
      ></span>
    </div>
    <div slot="workspace-controls" class="flex gap-x-2 items-center">
      {#if rcUrl}
        <Button
          type="secondary"
          small
          href={rcUrl}
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center gap-2"
        >
          <span>View in Cloud</span>
        </Button>
      {/if}
      <Tooltip distance={8}>
        <Button
          type="secondary"
          small
          onClick={() => (pullDialogOpen = true)}
          disabled={!isProjectLinked}
          class="flex items-center gap-2"
        >
          <Download size="14px" />
          <span>Pull</span>
        </Button>
        <TooltipContent slot="tooltip-content">
          {isProjectLinked
            ? "Pull variables from Rill Cloud"
            : "Deploy to Rill Cloud to sync variables"}
        </TooltipContent>
      </Tooltip>
      <Tooltip distance={8}>
        <Button
          type="secondary"
          small
          onClick={() => (pushDialogOpen = true)}
          disabled={!isProjectLinked}
          class="flex items-center gap-2"
        >
          <Upload size="14px" />
          <span>Push</span>
        </Button>
        <TooltipContent slot="tooltip-content">
          {isProjectLinked
            ? "Push variables to Rill Cloud"
            : "Deploy to Rill Cloud to sync variables"}
        </TooltipContent>
      </Tooltip>
      {#if viewMode === "viz"}
        <Button
          type="primary"
          small
          onClick={() => (addDialogOpen = true)}
          class="flex items-center gap-2"
        >
          <Plus size="14px" />
          <span>Add variable</span>
        </Button>
      {/if}
    </div>
  </FileWorkspaceHeader>

  <WorkspaceEditorContainer slot="body">
    {#if viewMode === "code"}
      <Editor
        {fileArtifact}
        {extensions}
        bind:editor
        bind:autoSave={$autoSave}
      />
    {:else}
      <div class="h-full w-full overflow-auto p-4 bg-surface-background">
        <EnvVariablesTable data={envVariables} {actionsColumn} />
      </div>
    {/if}
  </WorkspaceEditorContainer>
</WorkspaceContainer>

<AddEnvDialog
  bind:open={addDialogOpen}
  existingVariables={envVariables}
  on:add={handleAddVariables}
/>

<PullEnvDialog bind:open={pullDialogOpen} {isProjectLinked} />

<PushEnvDialog bind:open={pushDialogOpen} {isProjectLinked} />

<style lang="postcss">
  button {
    @apply flex-none flex items-center justify-center rounded-[4px];
    @apply size-[22px] cursor-pointer;
  }

  .toggle {
    @apply bg-surface-hover;
  }

  .radio {
    @apply h-fit bg-surface-subtle border p-0.5 rounded-[6px] flex;
  }
</style>
