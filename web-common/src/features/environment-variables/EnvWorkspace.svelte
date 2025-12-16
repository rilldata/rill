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
  import { Code2Icon, Plus, Settings } from "lucide-svelte";
  import { parse as parseDotenv } from "dotenv";
  import type { EditorView } from "@codemirror/view";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import { flexRender } from "@tanstack/svelte-table";
  import ActionsCell from "./ActionsCell.svelte";
  import AddEnvDialog from "./AddEnvDialog.svelte";

  export let fileArtifact: any;

  let editor: EditorView;
  let viewMode: "code" | "viz" = "viz";
  let addDialogOpen = false;

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
  }

  async function handleDeleteVariable(key: string) {
    const updatedVariables = envVariables.filter((v) => v.key !== key);
    await updateEnvFile(updatedVariables);
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
  <div slot="header" class="flex flex-col w-full">
    <FileWorkspaceHeader
      resource={undefined}
      resourceKind={undefined}
      filePath={path}
      hasUnsavedChanges={$hasUnsavedChanges}
    />
    <div
      class="flex items-center justify-between px-4 py-2 border-t border-gray-200"
    >
      <div class="flex items-center gap-2">
        <div class="radio relative">
          {#each [{ view: "code", icon: Code2Icon, label: "Code view" }, { view: "viz", icon: Settings, label: "No-code view" }] as { view, icon: Icon, label } (view)}
            <Tooltip activeDelay={700} distance={8}>
              <button
                aria-label="Switch to {label}"
                id="{view}-toggle"
                class="size-[22px] z-10 hover:brightness-75"
                on:click={handleToggleView}
              >
                <Icon
                  size="15px"
                  color={view === viewMode ? "#4F46E5" : "#9CA3AF"}
                />
              </button>
              <TooltipContent slot="tooltip-content">
                {label}
              </TooltipContent>
            </Tooltip>
          {/each}
          <span
            style:left={viewMode === "code" ? "2px" : "24px"}
            class="toggle size-[22px] pointer-events-none absolute rounded-[4px] z-0 transition-[left]"
          />
        </div>
      </div>
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
  </div>

  <WorkspaceEditorContainer slot="body">
    {#if viewMode === "code"}
      <Editor
        {fileArtifact}
        {extensions}
        bind:editor
        bind:autoSave={$autoSave}
      />
    {:else}
      <div class="h-full w-full overflow-auto p-4">
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

<style lang="postcss">
  button {
    @apply flex-none flex items-center justify-center rounded-[4px];
    @apply size-[22px] cursor-pointer;
  }

  .toggle {
    @apply bg-surface outline outline-slate-200 outline-[1px];
  }

  .radio {
    @apply h-fit bg-slate-100 p-[2px] rounded-[6px] flex;
  }
</style>
