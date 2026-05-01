<script lang="ts">
  import { page } from "$app/stores";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type {
    PathOption,
    PathOptions,
  } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import CanvasPreviewCTAs from "@rilldata/web-common/features/canvas/CanvasPreviewCTAs.svelte";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import { getBreadcrumbOptions } from "@rilldata/web-common/features/dashboards/dashboard-utils";
  import {
    useValidCanvases,
    useValidExplores,
  } from "@rilldata/web-common/features/dashboards/selectors.js";
  import DeployProjectCTA from "@rilldata/web-common/features/dashboards/workspace/DeployProjectCTA.svelte";
  import ExplorePreviewCTAs from "@rilldata/web-common/features/explores/ExplorePreviewCTAs.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import PreviewModeToggleButton from "@rilldata/web-common/layout/header/PreviewModeToggleButton.svelte";
  import { isDeployPage } from "@rilldata/web-common/layout/navigation/route-utils";
  import { previewModeLocked } from "@rilldata/web-common/layout/preview-mode-store";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import InputWithConfirm from "../components/forms/InputWithConfirm.svelte";
  import Tag from "../components/tag/Tag.svelte";
  import { fileArtifacts } from "../features/entity-management/file-artifacts";

  const { deploy, developerChat, stickyDashboardState } = featureFlags;
  const runtimeClient = useRuntimeClient();

  export let mode: string;

  $: ({
    params: { name: dashboardName },
    route,
  } = $page);

  $: onVizRoute = route.id?.includes("explore") || route.id?.includes("canvas");

  $: ({ unsavedFiles } = fileArtifacts);
  $: ({ size: unsavedFileCount } = $unsavedFiles);
  $: onDeployPage = isDeployPage($page);
  $: showDeployCTA = $deploy && !onDeployPage;
  $: showDeveloperChat = $developerChat && !onDeployPage;
  $: showPreviewToggle = !onDeployPage && !$previewModeLocked;

  $: exploresQuery = useValidExplores(runtimeClient);
  $: canvasQuery = useValidCanvases(runtimeClient);
  $: projectTitleQuery = useProjectTitle(runtimeClient);

  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";

  $: explores = $exploresQuery?.data ?? [];
  $: canvases = $canvasQuery?.data ?? [];

  // Resolve the dashboard the user is currently editing (if any) so the
  // Preview toggle can navigate directly to that dashboard's preview route
  // (and back to its file in Edit mode), instead of bouncing through the
  // /dashboards listing.
  $: editedFilePath = $page.url.pathname.startsWith("/files")
    ? $page.url.pathname.slice("/files".length)
    : null;

  $: editedExplore = editedFilePath
    ? explores.find((e) => e?.meta?.filePaths?.includes(editedFilePath))
    : null;
  $: editedCanvas =
    !editedExplore && editedFilePath
      ? canvases.find((c) => c?.meta?.filePaths?.includes(editedFilePath))
      : null;

  $: viewedExplore =
    mode === "Preview" && route.id?.includes("explore") && dashboardName
      ? explores.find((e) => e?.meta?.name?.name === dashboardName)
      : null;
  $: viewedCanvas =
    mode === "Preview" && route.id?.includes("canvas") && dashboardName
      ? canvases.find((c) => c?.meta?.name?.name === dashboardName)
      : null;

  $: previewToggleHref = (() => {
    if (mode === "Preview") {
      const filePath =
        viewedExplore?.meta?.filePaths?.[0] ??
        viewedCanvas?.meta?.filePaths?.[0];
      return filePath ? `/files${filePath}` : "/";
    }
    if (editedExplore) return `/explore/${editedExplore.meta?.name?.name}`;
    if (editedCanvas) return `/canvas/${editedCanvas.meta?.name?.name}`;
    return "/dashboards";
  })();

  $: defaultDashboard = explores[0] ?? canvases[0] ?? null;

  $: hasValidDashboard = Boolean(defaultDashboard);

  $: dashboardOptions = {
    options: getBreadcrumbOptions(explores, canvases),
    carryOverSearchParams: $stickyDashboardState,
  } satisfies PathOptions;

  $: projectPath = <PathOption>{
    label: projectTitle,
    section: "project",
    depth: -1,
    href: mode === "Preview" ? "/dashboards" : "/",
  };

  $: pathParts = [
    { options: new Map([[projectTitle.toLowerCase(), projectPath]]) },
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

<Header borderBottom={!onDeployPage && mode !== "Preview"}>
  {#if !onDeployPage}
    <HeaderLogo href={mode === "Preview" ? "/dashboards" : "/"} />

    <Tag text={mode} color="gray"></Tag>

    {#if mode === "Preview" || onVizRoute}
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
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if route.id?.includes("explore")}
      <ExplorePreviewCTAs exploreName={dashboardName} />
    {:else if route.id?.includes("canvas")}
      <CanvasPreviewCTAs canvasName={dashboardName} />
    {:else if showDeveloperChat}
      <ChatToggle />
    {/if}
    {#if showPreviewToggle}
      <PreviewModeToggleButton
        mode={mode === "Preview" ? "Edit" : "Preview"}
        href={previewToggleHref}
      />
    {/if}
    {#if showDeployCTA}
      <DeployProjectCTA {hasValidDashboard} />
    {/if}
    <LocalAvatarButton />
  </div>
</Header>
