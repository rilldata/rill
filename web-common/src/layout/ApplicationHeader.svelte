<script context="module" lang="ts">
  // Editing `/files/dashboards/<name>.yaml` previews the corresponding
  // explore/canvas dashboard directly; anywhere else the Preview button
  // falls back to the dashboards listing.
  const DASHBOARD_FILE_RE = /^\/files\/dashboards\/(.+)\.yaml$/;
  // Inverse: viewing `/explore/<name>` or `/canvas/<name>` jumps the Edit
  // button to that dashboard's yaml; anywhere else falls back to the editor home.
  const DASHBOARD_VIEW_RE = /^\/(?:explore|canvas)\/(.+)$/;

  function getPreviewUrl(
    pathname: string,
    explores: V1Resource[],
    canvases: V1Resource[],
  ): string {
    const match = pathname.match(DASHBOARD_FILE_RE);
    if (!match) return "/dashboards";
    const name = match[1];
    if (explores.some((e) => e.meta?.name?.name === name)) {
      return `/explore/${name}`;
    }
    if (canvases.some((c) => c.meta?.name?.name === name)) {
      return `/canvas/${name}`;
    }
    return "/dashboards";
  }

  function getEditUrl(pathname: string): string {
    const match = pathname.match(DASHBOARD_VIEW_RE);
    if (!match) return "/";
    return `/files/dashboards/${match[1]}.yaml`;
  }
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import type {
    PathOption,
    PathOptions,
  } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import { Button } from "@rilldata/web-common/components/button";
  import LocalAvatarButton from "@rilldata/web-common/features/authentication/LocalAvatarButton.svelte";
  import CanvasPreviewCTAs from "@rilldata/web-common/features/canvas/CanvasPreviewCTAs.svelte";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import { getBreadcrumbOptions } from "@rilldata/web-common/features/dashboards/dashboard-utils";
  import {
    useValidCanvases,
    useValidExplores,
  } from "@rilldata/web-common/features/dashboards/selectors.js";
  import ViewAsButton from "@rilldata/web-common/features/dashboards/granular-access-policies/ViewAsButton.svelte";
  import DeployProjectCTA from "@rilldata/web-common/features/dashboards/workspace/DeployProjectCTA.svelte";
  import ExplorePreviewCTAs from "@rilldata/web-common/features/explores/ExplorePreviewCTAs.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import ProjectTitleEditor from "@rilldata/web-common/features/project/ProjectTitleEditor.svelte";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import { isDeployPage } from "@rilldata/web-common/layout/navigation/route-utils";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import Tag from "../components/tag/Tag.svelte";

  const { deploy, developerChat, readOnly, stickyDashboardState } =
    featureFlags;
  const runtimeClient = useRuntimeClient();

  export let mode: string;

  $: ({
    params: { name: dashboardName },
    route,
  } = $page);

  $: onVizRoute = route.id?.includes("explore") || route.id?.includes("canvas");

  $: onDeployPage = isDeployPage($page);
  $: showDeployCTA = $deploy && !onDeployPage;
  $: showDeveloperChat = $developerChat && !onDeployPage;

  $: exploresQuery = useValidExplores(runtimeClient);
  $: canvasQuery = useValidCanvases(runtimeClient);
  $: projectTitleQuery = useProjectTitle(runtimeClient);

  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";

  $: explores = $exploresQuery?.data ?? [];
  $: canvases = $canvasQuery?.data ?? [];

  $: defaultDashboard = explores[0] ?? canvases[0] ?? null;

  $: hasValidDashboard = Boolean(defaultDashboard);

  $: previewUrl = getPreviewUrl($page.url.pathname, explores, canvases);
  $: editUrl = getEditUrl($page.url.pathname);

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
      <ProjectTitleEditor />
    {/if}
  {/if}

  <div class="flex gap-x-2 items-center ml-auto">
    {#if route.id?.includes("explore")}
      <ExplorePreviewCTAs exploreName={dashboardName} />
    {:else if route.id?.includes("canvas")}
      <CanvasPreviewCTAs canvasName={dashboardName} />
    {:else}
      {#if mode === "Preview" && !$readOnly}
        <ViewAsButton />
        <Button type="secondary" href={editUrl}>Edit</Button>
      {:else if mode === "Developer" && !$readOnly}
        <Button type="secondary" href={previewUrl}>Preview</Button>
      {/if}
      {#if showDeveloperChat && mode !== "Preview"}
        <ChatToggle />
      {/if}
    {/if}
    {#if showDeployCTA}
      <DeployProjectCTA {hasValidDashboard} />
    {/if}
    <LocalAvatarButton />
  </div>
</Header>
