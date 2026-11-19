<script lang="ts">
  import { goto } from "$app/navigation";
  import { customYAMLwithJSONandSQL } from "@rilldata/web-common/components/editor/presets/yamlWithJsonAndSql";
  import Editor from "@rilldata/web-common/features/editor/Editor.svelte";
  import type { EditorView } from "@codemirror/view";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/actions/ui-actions.ts";
  import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ReconcileWarningPanel from "@rilldata/web-common/features/entity-management/ReconcileWarningPanel.svelte";
  import { sendComponentFilePrompt } from "@rilldata/web-common/features/custom-viz/component-ai-agent";
  import ComponentPreview from "@rilldata/web-common/features/custom-viz/workspace/ComponentPreview.svelte";
  import EjectButton from "@rilldata/web-common/features/custom-viz/workspace/EjectButton.svelte";
  import TestBindingsPanel from "@rilldata/web-common/features/custom-viz/workspace/TestBindingsPanel.svelte";
  import UsedByPanel from "@rilldata/web-common/features/custom-viz/workspace/UsedByPanel.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { getPreviewArgsStore } from "@rilldata/web-common/features/custom-viz/workspace/preview-state";
  import { getDeclaredParams } from "@rilldata/web-common/features/custom-viz/params";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import {
    Inspector,
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import WorkspaceEditorContainer from "@rilldata/web-common/layout/workspace/WorkspaceEditorContainer.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1ComponentParam } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let fileArtifact: FileArtifact;

  const runtimeClient = useRuntimeClient();

  const { developerChat } = featureFlags;

  let aiPrompt = "";
  let editor: EditorView | null = null;

  function askAI() {
    const prompt = aiPrompt.trim();
    if (!prompt) return;
    sendComponentFilePrompt(runtimeClient, filePath, prompt);
    aiPrompt = "";
  }

  $: ({
    autoSave,
    path: filePath,
    fileName,
    getResource,
    remoteContent,
    hasUnsavedChanges,
  } = fileArtifact);

  $: resourceQuery = getResource(queryClient);
  $: ({ data: resource } = $resourceQuery);

  $: componentName = getNameFromFile(filePath);

  $: workspace = workspaces.get(filePath);
  $: selectedViewStore = workspace.view;
  $: selectedView = $selectedViewStore ?? "code";

  $: parseErrorQuery = fileArtifact.getParseError(queryClient);
  $: parseError = $parseErrorQuery;

  $: argsStore = getPreviewArgsStore(filePath);

  // Breaking-change warning: the component is used by dashboards, and the draft
  // spec removes, retypes, or adds a required param compared to the valid spec.
  $: canvasesQuery = useFilteredResources(runtimeClient, ResourceKind.Canvas);
  $: usedByCount = ($canvasesQuery?.data ?? []).filter((res) =>
    (res.meta?.refs ?? []).some(
      (ref) =>
        ref.kind === ResourceKind.Component && ref.name === componentName,
    ),
  ).length;
  $: validParams = resource?.component?.state?.validSpec?.params ?? [];
  $: draftParams = getDeclaredParams(resource, true);
  $: showBreakingWarning =
    usedByCount > 0 &&
    validParams.length > 0 &&
    isBreakingChange(validParams, draftParams);

  function isBreakingChange(
    valid: V1ComponentParam[],
    draft: V1ComponentParam[],
  ): boolean {
    const draftByName = new Map(draft.map((p) => [p.name, p]));
    for (const param of valid) {
      const draftParam = draftByName.get(param.name);
      if (!draftParam || draftParam.type !== param.type) return true;
    }
    for (const param of draft) {
      const isNew = !valid.some((v) => v.name === param.name);
      if (isNew && param.required) return true;
    }
    return false;
  }

  async function onChangeCallback(newTitle: string) {
    const newRoute = await handleEntityRename(
      runtimeClient,
      newTitle,
      filePath,
      fileName,
    );
    if (newRoute) await goto(newRoute);
  }
</script>

<WorkspaceContainer>
  <WorkspaceHeader
    slot="header"
    {filePath}
    {resource}
    hasUnsavedChanges={$hasUnsavedChanges}
    titleInput={fileName}
    codeToggle
    onTitleChange={onChangeCallback}
    resourceKind={ResourceKind.Component}
  >
    {#snippet cta()}
      <EjectButton
        {fileArtifact}
        {componentName}
        {resource}
        args={$argsStore}
      />
    {/snippet}
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    <div class="flex flex-col h-full">
      {#if showBreakingWarning}
        <div
          class="px-4 py-2 text-xs bg-yellow-50 text-yellow-800 border-b border-yellow-200"
        >
          {m.component_breaking_warning({ count: String(usedByCount) })}
        </div>
      {/if}
      <div class="flex-1 min-h-0">
        <!-- showError in both views so the error banner (with its AI fix button)
             is visible from the preview too, not just the code editor. -->
        <WorkspaceEditorContainer
          {resource}
          {parseError}
          remoteContent={$remoteContent}
          {filePath}
        >
          {#if selectedView === "code"}
            <Editor
              bind:editor
              {fileArtifact}
              extensions={[customYAMLwithJSONandSQL]}
              bind:autoSave={$autoSave}
            />
          {:else}
            <ComponentPreview
              {componentName}
              {resource}
              args={$argsStore}
              {filePath}
            />
          {/if}
        </WorkspaceEditorContainer>
      </div>
      <ReconcileWarningPanel {fileArtifact} />
    </div>
  </svelte:fragment>

  <svelte:fragment slot="inspector">
    <Inspector minWidth={320} maxWidth={480} {filePath}>
      <SidebarWrapper
        type="secondary"
        disableHorizontalPadding
        title={m.component_preview_title()}
      >
        {#if $developerChat}
          <div class="flex flex-col gap-y-1 px-5 py-3 border-b">
            <textarea
              class="w-full p-2 text-sm border border-gray-300 rounded-sm"
              rows="2"
              placeholder={m.component_ai_placeholder()}
              bind:value={aiPrompt}
              onkeydown={(event) => {
                if (event.key === "Enter" && !event.shiftKey) {
                  event.preventDefault();
                  askAI();
                }
              }}
            ></textarea>
          </div>
        {/if}
        <TestBindingsPanel {resource} {argsStore} />
        <UsedByPanel {componentName} {argsStore} />
      </SidebarWrapper>
    </Inspector>
  </svelte:fragment>
</WorkspaceContainer>
