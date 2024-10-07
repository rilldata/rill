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
  import { ResourceKind } from "../features/entity-management/resource-selectors";
  import type { V1Resource, V1ResourceName } from "../runtime-client";
  import type { FileArtifact } from "../features/entity-management/file-artifact";
  import PreviewButton from "../features/explores/PreviewButton.svelte";

  export let mode: string;

  $: ({ instanceId } = $runtime);

  $: ({
    params: { name: dashboardName },
    route,
    data,
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

  $: fileArtifact = data.fileArtifact as FileArtifact | undefined;

  $: resourceNameStore = fileArtifact?.resourceName;
  $: remoteContent = fileArtifact?.remoteContent;
  $: inferredResourceKind = fileArtifact?.inferredResourceKind;

  $: isNewMetricsView = $remoteContent?.includes("version: 1");

  $: dashboardResourceInErrorState = Boolean(
    $inferredResourceKind &&
      getPreviewType($inferredResourceKind) !== null &&
      !isNewMetricsView &&
      previewHref === null,
  );

  $: resourceName = resourceNameStore && $resourceNameStore;

  $: previewType = getPreviewType(resourceName?.kind);
  $: previewName = getPreviewName(resourceName, explores);
  $: previewHref =
    previewName && previewType ? `/${previewType}/${previewName}` : null;

  $: defaultType = getPreviewType(defaultDashboard?.meta?.name?.kind);
  $: defaultName = getPreviewName(defaultDashboard?.meta?.name, explores);
  $: fallbackHref =
    defaultName && defaultType ? `/${defaultType}/${defaultName}` : null;

  $: href =
    previewType === null || isNewMetricsView ? fallbackHref : previewHref;

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

    parsed.set("title", editedTitle);

    artifact.updateLocalContent(parsed.toString(), true);
    await artifact.saveLocalContent();
  }

  function getPreviewType(
    kind: string | undefined,
  ): "custom" | "explore" | null {
    switch (kind) {
      case ResourceKind.Canvas:
        return "custom";
      case ResourceKind.Explore:
      case ResourceKind.MetricsView:
        return "explore";
      default:
        return null;
    }
  }

  function getPreviewName(
    name: V1ResourceName | undefined,
    validExploreDashboards: V1Resource[],
  ): string | null {
    switch (name?.kind) {
      case ResourceKind.Canvas:
        return name?.name ?? null;
      case ResourceKind.Explore:
        return (
          validExploreDashboards.find(
            ({ meta }) => meta?.name?.name === name.name,
          )?.meta?.name?.name ?? null
        );
      case ResourceKind.MetricsView:
        return (
          validExploreDashboards.find(
            ({ explore }) => explore?.spec?.metricsView === name.name,
          )?.meta?.name?.name ?? null
        );
      default:
        return null;
    }
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
    {:else if mode === "Developer"}
      <PreviewButton {href} disabled={!href || dashboardResourceInErrorState} />
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
