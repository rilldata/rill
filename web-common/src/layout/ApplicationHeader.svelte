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
  import DeployProjectCTA from "@rilldata/web-common/features/dashboards/workspace/DeployProjectCTA.svelte";
  import ExplorePreviewCTAs from "@rilldata/web-common/features/explores/ExplorePreviewCTAs.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { useProjectTitle } from "@rilldata/web-common/features/project/selectors";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import { isDeployPage } from "@rilldata/web-common/layout/navigation/route-utils";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { get } from "svelte/store";
  import { parseDocument } from "yaml";
  import InputWithConfirm from "../components/forms/InputWithConfirm.svelte";
  import Tag from "../components/tag/Tag.svelte";
  import { fileArtifacts } from "../features/entity-management/file-artifacts";

  const { deploy, developerChat, readOnly, stickyDashboardState } =
    featureFlags;
  const runtimeClient = useRuntimeClient();

  export let mode: string;
  export let onModeToggle: (() => void) | undefined = undefined;
  export let modeLocked: boolean = false;

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

  $: exploresQuery = useValidExplores(runtimeClient);
  $: canvasQuery = useValidCanvases(runtimeClient);
  $: projectTitleQuery = useProjectTitle(runtimeClient);

  $: projectTitle = $projectTitleQuery?.data ?? "Untitled Rill Project";

  $: explores = $exploresQuery?.data ?? [];
  $: canvases = $canvasQuery?.data ?? [];

  $: defaultDashboard = explores[0] ?? canvases[0] ?? null;

  $: hasValidDashboard = Boolean(defaultDashboard);

  // When editing a dashboard YAML in /files/dashboards/[name].yaml, the
  // Preview button jumps directly to the corresponding explore or canvas.
  // Anywhere else in the file editor, it falls back to the dashboards
  // listing.
  $: previewUrl = (() => {
    const match = $page.url.pathname.match(/^\/files\/dashboards\/(.+)\.yaml$/);
    if (!match) return "/dashboards";
    const name = match[1];
    if (explores.some((e) => e.meta?.name?.name === name)) {
      return `/explore/${name}`;
    }
    if (canvases.some((c) => c.meta?.name?.name === name)) {
      return `/canvas/${name}`;
    }
    return "/dashboards";
  })();

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

    {#if onModeToggle && !modeLocked}
      <button
        type="button"
        class="contents cursor-pointer"
        title="Switch to {mode === 'Preview' ? 'Developer' : 'Preview'}"
        on:click={onModeToggle}
      >
        <Tag text={mode} color="gray"></Tag>
      </button>
    {:else}
      <Tag text={mode} color="gray"></Tag>
    {/if}

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
    {:else}
      {#if mode === "Preview" && !$readOnly}
        <Button type="secondary" href="/">Edit</Button>
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
