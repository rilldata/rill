<script context="module" lang="ts">
  // Editing `/-/edit/files/dashboards/<name>.yaml` previews the corresponding
  // explore/canvas dashboard directly; anywhere else the Preview button
  // falls back to the dashboards listing.
  const DASHBOARD_FILE_RE =
    /\/-\/edit\/files\/dashboards\/(.+)\.yaml(?:$|[/?#])/;

  function getEditPreviewSubpath(
    pathname: string,
    dashboards: V1Resource[],
  ): string {
    const match = pathname.match(DASHBOARD_FILE_RE);
    if (!match) return "/dashboards";
    const name = match[1];
    const resource = dashboards.find((r) => r.meta?.name?.name === name);
    if (resource?.explore) return `/explore/${name}`;
    if (resource?.canvas) return `/canvas/${name}`;
    return "/dashboards";
  }
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { isEditPreviewRoute } from "./edit-route-utils";

  export let organization: string;
  export let project: string;
  export let branch: string;

  const runtimeClient = useRuntimeClient();
  $: dashboardsQuery = useDashboards(runtimeClient);
  $: dashboards = $dashboardsQuery.data ?? [];

  $: editPrefix = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  $: inPreview = isEditPreviewRoute($page.url.pathname);
  $: previewSubpath = getEditPreviewSubpath($page.url.pathname, dashboards);
  $: navHref = inPreview ? editPrefix : `${editPrefix}${previewSubpath}`;
  $: navLabel = inPreview ? "Edit" : "Preview";
  $: navTooltip = inPreview
    ? "Switch to developer mode"
    : "Switch to preview mode";
</script>

<Tooltip distance={8}>
  <Button type="secondary" href={navHref} class="!bg-surface-base">
    {navLabel}
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">{navTooltip}</span>
  </TooltipContent>
</Tooltip>
