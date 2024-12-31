<script lang="ts">
  import { page } from "$app/stores";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import { getBreadcrumbOptions } from "@rilldata/web-common/features/dashboards/dashboard-utils";
  import {
    useValidCanvases,
    useValidExplores,
  } from "@rilldata/web-common/features/dashboards/selectors.js";
  import DeployProjectCTA from "@rilldata/web-common/features/dashboards/workspace/DeployProjectCTA.svelte";
  import ExplorePreviewCTAs from "@rilldata/web-common/features/explores/ExplorePreviewCTAs.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { isDeployPage } from "@rilldata/web-common/layout/navigation/route-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import InputWithConfirm from "../components/forms/InputWithConfirm.svelte";
  import { fileArtifacts } from "../features/entity-management/file-artifacts";

  export let mode: string;

  $: ({ instanceId } = $runtime);

  $: ({
    params: { name: dashboardName },
    route,
  } = $page);

  $: ({ unsavedFiles } = fileArtifacts);
  $: ({ size: unsavedFileCount } = $unsavedFiles);
  $: onDeployPage = isDeployPage($page);

  $: exploresQuery = useValidExplores(instanceId);
  $: canvasQuery = useValidCanvases(instanceId);
  $: projectTitleQuery = useProjectTitle(instanceId);

  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";

  $: explores = $exploresQuery?.data ?? [];
  $: canvases = $canvasQuery?.data ?? [];

  $: defaultDashboard = explores[0] ?? canvases[0] ?? null;

  $: hasValidDashboard = Boolean(defaultDashboard);

  $: dashboardOptions = getBreadcrumbOptions(explores, canvases);

  $: projectPath = <PathOption>{
    label: projectTitle,
    section: "project",
    depth: -1,
    href: "/",
  };

  $: pathParts = [
    new Map([[projectTitle.toLowerCase(), projectPath]]),
    dashboardOptions,
  ];

  $: currentPath = [projectTitle, dashboardName?.toLowerCase()];

  async function submitTitleChange(editedTitle: string) {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");

    let content = get(artifact.editorContent);

    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.remoteContent);
      if (!content) {
        return;
      }
    }
    const parsed = parseDocument(content);

    parsed.set("display_name", editedTitle);

    artifact.updateEditorContent(parsed.toString(), true);
    await artifact.saveLocalContent();
  }
</script>

<header>
  <a href="/">
    <Rill />
  </a>

  <span class="rounded-full px-2 border text-gray-800 bg-gray-50">
    {mode}
  </span>

  {#if mode === "Preview"}
    {#if $exploresQuery?.data}
      <Breadcrumbs {pathParts} {currentPath} />
    {/if}
  {:else if mode === "Developer"}
    <InputWithConfirm
      size="md"
      bumpDown
      type="Project"
      textClass="font-medium"
      value={projectTitle}
      onConfirm={submitTitleChange}
      showIndicator={unsavedFileCount > 0}
    />
  {/if}

  <div class="ml-auto flex gap-x-2">
    {#if mode === "Preview"}
      {#if route.id?.includes("explore")}
        <ExplorePreviewCTAs exploreName={dashboardName} />
      {/if}
    {/if}
    {#if !onDeployPage}
      <DeployProjectCTA {hasValidDashboard} />
    {/if}
    <LocalAvatarButton />
  </div>
</header>

<style lang="postcss">
  header {
    @apply w-full bg-background box-border;
    @apply flex gap-x-2 items-center px-4 border-b flex-none;
    @apply h-11;
  }
</style>
