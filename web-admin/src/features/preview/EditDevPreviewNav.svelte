<script lang="ts">
  import { page } from "$app/stores";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import {
    createRuntimeServiceGetResource,
    createRuntimeServiceListResources,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { tick } from "svelte";
  import { branchPathPrefix } from "../branches/branch-utils";

  const { chat } = featureFlags;
  const runtimeClient = useRuntimeClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: branch = $page.url.pathname.match(/\/@([^/]+)/)?.[1];
  $: editPrefix = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;

  $: currentPath = $page.url.pathname;

  $: baseTabs = [
    { id: "dashboards", label: "Dashboards", path: `${editPrefix}/dashboards` },
    { id: "status", label: "Status", path: `${editPrefix}/status` },
  ];

  $: aiTab = { id: "ai", label: "AI", path: `${editPrefix}/ai` };

  $: tabs = $chat ? [baseTabs[0], aiTab, baseTabs[1]] : baseTabs;

  $: activeTab =
    tabs.find((t) => currentPath.startsWith(t.path))?.id ?? "dashboards";

  $: selectedIndex = tabs.findIndex((t) => t.id === activeTab);

  let tabElements: HTMLAnchorElement[] = [];
  let indicatorLeft = 0;
  let indicatorWidth = 0;

  function updateIndicator() {
    const el = tabElements[selectedIndex];
    if (el) {
      indicatorLeft = el.offsetLeft;
      indicatorWidth = el.clientWidth;
    }
  }

  $: if (selectedIndex >= 0 && tabElements.length) {
    void tick().then(updateIndicator);
  }

  // Project status indicator (mirrors local's LocalProjectStatusIndicator)
  $: hasResourceErrorsQuery = createRuntimeServiceListResources(
    runtimeClient,
    {},
    {
      query: {
        select: (data) =>
          (data.resources ?? []).filter((r) => !!r.meta?.reconcileError)
            .length > 0,
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  $: projectParserQuery = createRuntimeServiceGetResource(
    runtimeClient,
    {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    },
    {
      query: {
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  $: hasResourceErrors = $hasResourceErrorsQuery.data;
  $: hasParseErrors =
    ($projectParserQuery.data?.resource?.projectParser?.state?.parseErrors
      ?.length ?? 0) > 0;
  $: statusLoading =
    $hasResourceErrorsQuery.isLoading || $projectParserQuery.isLoading;
  $: statusErrored =
    $hasResourceErrorsQuery.isError || $projectParserQuery.isError;
</script>

<svelte:window onresize={updateIndicator} />

<div class="nav-bar">
  <nav>
    {#each tabs as tab, i (tab.id)}
      <a
        href={tab.path}
        class="tab"
        class:selected={activeTab === tab.id}
        bind:this={tabElements[i]}
      >
        <p data-content={tab.label}>{tab.label}</p>
        {#if tab.id === "status"}
          {#if statusLoading}
            <LoadingSpinner />
          {:else if statusErrored || hasResourceErrors || hasParseErrors}
            <CancelCircle className="text-red-600" />
          {:else}
            <CheckCircle className="text-green-400" />
          {/if}
        {/if}
      </a>
    {/each}
  </nav>

  {#if indicatorWidth > 0}
    <span
      class="indicator"
      style:width="{indicatorWidth}px"
      style:transform="translateX({indicatorLeft}px)"
    ></span>
  {/if}
</div>

<style lang="postcss">
  .nav-bar {
    @apply bg-surface-base border-b pt-1;
    @apply gap-y-[3px] flex flex-col;
  }

  nav {
    @apply flex w-fit;
    @apply gap-x-3 px-[17px];
  }

  .tab {
    @apply px-2 py-1.5 flex gap-x-1 items-center w-fit;
    @apply rounded-sm text-fg-muted;
    @apply text-xs font-medium justify-center;
    @apply no-underline;
  }

  .selected {
    @apply text-fg-accent font-semibold;
  }

  .tab:hover {
    background: var(--surface-hover);
  }

  .tab p {
    @apply text-center;
  }

  .tab p::before {
    content: attr(data-content);
    display: block;
    font-weight: 600;
    height: 0;
    visibility: hidden;
  }

  .indicator {
    @apply h-[3px] bg-primary-500 rounded;
    transition:
      transform 0.2s ease,
      width 0.2s ease;
  }
</style>
