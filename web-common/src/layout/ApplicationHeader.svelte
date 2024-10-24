<script lang="ts">
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import DeployProjectCTA from "@rilldata/web-common/features/dashboards/workspace/DeployProjectCTA.svelte";
  import { page } from "$app/stores";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { getBreadcrumbOptions } from "@rilldata/web-common/features/dashboards/dashboard-utils";
  import {
    useValidCanvases,
    useValidExplores,
  } from "@rilldata/web-common/features/dashboards/selectors.js";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ExplorePreviewCTAs from "@rilldata/web-common/features/explores/ExplorePreviewCTAs.svelte";
  import InputWithConfirm from "../components/forms/InputWithConfirm.svelte";
  import { get } from "svelte/store";
  import { fileArtifacts } from "../features/entity-management/file-artifacts";
  import { parseDocument } from "yaml";

  export let mode: string;

  $: ({ instanceId } = $runtime);

  $: ({
    params: { name: dashboardName },
    route,
  } = $page);

  $: ({ unsavedFiles } = fileArtifacts);
  $: ({ size: unsavedFileCount } = $unsavedFiles);

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

  $: currentDashboard = explores.find(
    (d) => d.meta?.name?.name?.toLowerCase() === dashboardName?.toLowerCase(),
  );

  $: metricsViewName = currentDashboard?.meta?.name?.name;

  async function submitTitleChange(editedTitle: string) {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");

    let content = get(artifact.localContent) ?? get(artifact.remoteContent);

    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.localContent) ?? get(artifact.remoteContent);
      if (!content) {
        return;
      }
    }
    const parsed = parseDocument(content);

    parsed.set("display_name", editedTitle);

    artifact.updateLocalContent(parsed.toString(), true);
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
      {#if route.id?.includes("explore") && metricsViewName}
        <StateManagersProvider {metricsViewName} exploreName={dashboardName}>
          <ExplorePreviewCTAs exploreName={dashboardName} />
        </StateManagersProvider>
      {/if}
    {/if}
    <DeployProjectCTA {hasValidDashboard} />
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
