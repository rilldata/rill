<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { errorStore } from "@rilldata/web-admin/components/errors/error-store";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import ViewAsUserPopover from "@rilldata/web-admin/features/view-as-user/ViewAsUserPopover.svelte";
  import { viewAsUserStore } from "@rilldata/web-admin/features/view-as-user/viewAsUserStore";
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import BreadcrumbItem from "@rilldata/web-common/components/navigation/breadcrumbs/BreadcrumbItem.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import type { PathOption } from "@rilldata/web-common/components/navigation/breadcrumbs/types";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    GitBranchIcon,
    GitPullRequestIcon,
    LinkIcon,
    XIcon,
  } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  export let activeBranch: string;
  export let canManageProject = false;

  const runtimeClient = useRuntimeClient();
  let viewAsOpen = false;

  $: branchPrefix = branchPathPrefix(activeBranch);
  $: branchHref = `/${organization}/${project}${branchPrefix}`;
  $: editHref = `${branchHref}/-/edit`;

  // Route shape: `/{org}/{project}/@{branch}/explore/[name]`,
  // `.../@{branch}/canvas/[name]`, or `.../@{branch}/dashboards`.
  $: dashboardName = $page.params.name as string | undefined;
  $: routeId = $page.route.id ?? "";
  $: onExplore = routeId.includes("/explore/[name]");
  $: onCanvas = routeId.includes("/canvas/[name]");
  $: onDashboardPage = onExplore || onCanvas;

  // Dashboard switcher dropdown — same approach the editor uses.
  $: dashboardsQuery = useDashboards(runtimeClient);
  $: dashboards = $dashboardsQuery.data ?? [];
  $: dashboardOptions = {
    options: [...dashboards]
      .sort((a, b) => {
        const aIsCanvas = !!a?.canvas;
        const bIsCanvas = !!b?.canvas;
        if (aIsCanvas !== bIsCanvas) return aIsCanvas ? -1 : 1;
        return (a.meta?.name?.name ?? "").localeCompare(
          b.meta?.name?.name ?? "",
        );
      })
      .reduce((map, resource) => {
        const name = resource.meta?.name?.name ?? "";
        const isExplore = !!resource?.explore;
        const section = isExplore ? "explore" : "canvas";
        const label =
          (isExplore
            ? resource?.explore?.spec?.displayName
            : resource?.canvas?.spec?.displayName) || name;
        return map.set(name.toLowerCase(), {
          label,
          href: `${branchHref}/${section}/${name}`,
          resourceKind: isExplore ? ResourceKind.Explore : ResourceKind.Canvas,
        });
      }, new Map<string, PathOption>()),
  };

  // "Back to edit" lands on the YAML for the dashboard the user is previewing,
  // or on the editor root if they're on the dashboards listing.
  $: backToEditHref = dashboardName
    ? `${editHref}/files/dashboards/${dashboardName}.yaml`
    : editHref;

  $: shareableUrl = `${$page.url.origin}${branchHref}${$page.url.pathname.startsWith(branchHref) ? $page.url.pathname.slice(branchHref.length) : ""}${$page.url.search}`;

  function copyShareLink() {
    copyToClipboard(shareableUrl, "Shareable URL has been copied.");
  }

  function clearViewAs(event: MouseEvent) {
    // Stop the click from bubbling into the dropdown trigger.
    event.stopPropagation();
    event.preventDefault();
    viewAsUserStore.set(null);
    errorStore.reset();
  }
</script>

<header class="preview-header">
  <div class="flex items-center min-w-0 gap-x-2">
    <span class="preview-pill">Preview</span>
    <nav class="flex items-center gap-x-2 min-w-0">
      <span class="text-fg-muted text-sm truncate">{project}</span>
      <Slash />
      <span
        class="text-fg-primary font-medium flex flex-row items-center gap-x-2"
      >
        <GitBranchIcon size="14" class="text-fg-primary" />
        <span class="truncate max-w-[200px]" title={activeBranch}>
          {activeBranch}
        </span>
      </span>
    </nav>
  </div>

  {#if onDashboardPage && dashboardName}
    <div class="flex items-center justify-center px-3">
      <BreadcrumbItem
        depth={0}
        pathOptions={dashboardOptions}
        current={dashboardName.toLowerCase()}
        isCurrentPage={true}
      />
    </div>
  {/if}

  <div class="flex items-center gap-x-2 ml-auto">
    {#if canManageProject}
      <DropdownMenu.Root bind:open={viewAsOpen}>
        <DropdownMenu.Trigger>
          {#snippet child({ props })}
            <button {...props} class="view-as-pill" type="button">
              {#if $viewAsUserStore}
                <button
                  type="button"
                  class="view-as-clear"
                  aria-label="Clear view as"
                  onclick={clearViewAs}
                >
                  <XIcon size="14" />
                </button>
                <span class="font-normal">View as</span>
                <span class="font-semibold text-accent-primary-action">
                  {$viewAsUserStore.email}
                </span>
              {:else}
                <span class="font-normal">View as</span>
              {/if}
            </button>
          {/snippet}
        </DropdownMenu.Trigger>
        <DropdownMenu.Content
          align="end"
          class="flex flex-col min-w-[220px] max-w-[300px]"
        >
          <ViewAsUserPopover
            {organization}
            {project}
            onSelectUser={() => (viewAsOpen = false)}
          />
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    {/if}

    <Tooltip distance={8}>
      <button class="share-link" type="button" onclick={copyShareLink}>
        <LinkIcon size="14" />
        Share
      </button>
      <TooltipContent slot="tooltip-content" maxWidth="240px">
        <span class="text-xs"
          >Copy this branch's shareable cloud preview link</span
        >
      </TooltipContent>
    </Tooltip>

    <Tooltip distance={8}>
      <Button type="primary" disabled>
        <GitPullRequestIcon size="14" />
        Merge PR
      </Button>
      <TooltipContent slot="tooltip-content" maxWidth="200px">
        <span class="text-xs">Coming soon</span>
      </TooltipContent>
    </Tooltip>

    <Button type="secondary" href={backToEditHref}>Back to edit</Button>
  </div>
</header>

<style lang="postcss">
  .preview-header {
    @apply flex items-center px-4 h-12;
    @apply bg-surface-active border-b border-accent-primary;
  }

  .preview-pill {
    @apply inline-flex items-center h-7 px-2.5 rounded-2xl shrink-0;
    @apply border border-border bg-surface-base shadow-sm;
    @apply text-fg-secondary text-sm font-medium;
  }

  .view-as-pill {
    @apply inline-flex items-center gap-x-1 h-7 px-4 rounded-2xl;
    @apply border border-ring-focus bg-dimension text-fg-primary text-xs;
    @apply transition-colors;
  }

  .view-as-pill:hover {
    @apply bg-dimension/80;
  }

  /* Inner X button — picks up its own pointer events so clicks don't
     just open the dropdown. */
  .view-as-clear {
    @apply flex items-center justify-center rounded-sm size-4 -ml-1;
    @apply text-fg-secondary opacity-70 hover:opacity-100 hover:bg-surface-hover;
  }

  .share-link {
    @apply inline-flex items-center gap-x-2 h-7 px-3 rounded-sm;
    @apply text-accent-primary-action text-xs font-medium;
    @apply transition-colors;
  }

  .share-link:hover {
    @apply bg-surface-hover;
  }
</style>
