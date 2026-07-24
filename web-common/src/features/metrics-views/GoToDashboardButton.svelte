<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import CanvasIcon from "@rilldata/web-common/components/icons/CanvasIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    useGetCanvasesForMetricsView,
    useGetExploresForMetricsView,
  } from "../dashboards/selectors";
  import { allowPrimary } from "../dashboards/workspace/DeployProjectCTA.svelte";
  import {
    createCanvasDashboardFromMetricsView,
    createCanvasDashboardFromMetricsViewWithAgent,
  } from "./ai-generation/generateMetricsView";
  import { enableInlineExploreAndPreview } from "./inline-explore";
  import NavigateOrDropdown from "./NavigateOrDropdown.svelte";

  export let resource: V1Resource | undefined;
  // True when the metrics view file already defines an inline explore. The explore
  // is then reachable via the workspace's explore view and Preview button, so this
  // CTA only deals in canvases: it lists canvas dashboards instead of explores and
  // hides "Create Explore Dashboard".
  export let hasInlineExplore = false;

  // When a dashboard is defined inline in a metrics view file, its file path is the
  // metrics view file itself; the ?view=explore param opens the workspace on the
  // explore view.
  function dashboardHref(dashboard: V1Resource): string {
    const filePath = `/${removeLeadingSlash(dashboard?.meta?.filePaths?.[0] ?? "")}`;
    const definedInMetricsView =
      dashboard?.explore?.state?.validSpec?.definedInMetricsView;
    return getFileHref(filePath, definedInMetricsView ? "explore" : undefined);
  }

  const runtimeClient = useRuntimeClient();
  const { ai, developerChat } = featureFlags;

  $: dashboardsQuery = hasInlineExplore
    ? useGetCanvasesForMetricsView(
        runtimeClient,
        resource?.meta?.name?.name ?? "",
      )
    : useGetExploresForMetricsView(
        runtimeClient,
        resource?.meta?.name?.name ?? "",
      );
  $: dashboards = $dashboardsQuery.data ?? [];

  $: metricsViewFilePath = resource?.meta?.filePaths?.[0];
</script>

{#if dashboards?.length === 0}
  <div class="flex gap-2">
    <Button
      type="secondary"
      disabled={!resource}
      onClick={async () => {
        if (resource?.meta?.name?.name) {
          // Use developer agent if enabled, otherwise fall back to RPC
          if ($developerChat) {
            createCanvasDashboardFromMetricsViewWithAgent(
              runtimeClient,
              resource.meta.name.name,
            );
          } else {
            await createCanvasDashboardFromMetricsView(
              runtimeClient,
              resource.meta.name.name,
            );
          }
        }
      }}
    >
      Generate Canvas Dashboard{$ai ? " with AI" : ""}
    </Button>
    {#if !hasInlineExplore}
      <Button
        type={$allowPrimary ? "primary" : "secondary"}
        disabled={!resource}
        onClick={async () => {
          if (metricsViewFilePath)
            await enableInlineExploreAndPreview(
              queryClient,
              metricsViewFilePath,
            );
        }}
      >
        Create Explore Dashboard
      </Button>
    {/if}
  </div>
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <NavigateOrDropdown
          {...props}
          resources={dashboards}
          hrefForResource={dashboardHref}
        />
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end">
      <DropdownMenu.Group>
        {#each dashboards as resource (resource?.meta?.name?.name)}
          {@const isCanvas = resource?.meta?.name?.kind === ResourceKind.Canvas}
          {@const label =
            (isCanvas
              ? resource?.canvas?.state?.validSpec?.displayName
              : resource?.explore?.state?.validSpec?.displayName) ||
            resource?.meta?.name?.name}
          {@const filePath = resource?.meta?.filePaths?.[0]}
          {#if label && filePath}
            <DropdownMenu.Item href={dashboardHref(resource)}>
              {#if isCanvas}
                <CanvasIcon />
              {:else}
                <ExploreIcon />
              {/if}
              {label}
            </DropdownMenu.Item>
          {/if}
        {/each}
        <DropdownMenu.Separator />
        <DropdownMenu.Item
          onclick={async () => {
            if (resource?.meta?.name?.name) {
              // Use developer agent if enabled, otherwise fall back to RPC
              if ($developerChat) {
                createCanvasDashboardFromMetricsViewWithAgent(
                  runtimeClient,
                  resource.meta.name.name,
                );
              } else {
                await createCanvasDashboardFromMetricsView(
                  runtimeClient,
                  resource.meta.name.name,
                );
              }
            }
          }}
        >
          <Add />
          Generate Canvas Dashboard{$ai ? " with AI" : ""}
        </DropdownMenu.Item>
        {#if !hasInlineExplore}
          <DropdownMenu.Item
            onclick={async () => {
              if (metricsViewFilePath)
                await enableInlineExploreAndPreview(
                  queryClient,
                  metricsViewFilePath,
                );
            }}
          >
            <Add />
            Create Explore Dashboard
          </DropdownMenu.Item>
        {/if}
      </DropdownMenu.Group>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
