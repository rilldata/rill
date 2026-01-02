<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useGetExploresForMetricsView } from "../dashboards/selectors";
  import { allowPrimary } from "../dashboards/workspace/DeployProjectCTA.svelte";
  import { resourceColorMapping } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { createCanvasDashboardFromMetricsView } from "./ai-generation/generateMetricsView";
  import { createAndPreviewExplore } from "./create-and-preview-explore";
  import NavigateOrDropdown from "./NavigateOrDropdown.svelte";

  export let resource: V1Resource | undefined;

  const { ai, generateCanvas } = featureFlags;

  $: ({ instanceId } = $runtime);
  $: dashboardsQuery = useGetExploresForMetricsView(
    instanceId,
    resource?.meta?.name?.name ?? "",
  );
  $: dashboards = $dashboardsQuery.data ?? [];
</script>

{#if dashboards?.length === 0}
  <div class="flex gap-2">
    {#if $generateCanvas}
      <Button
        type="secondary"
        disabled={!resource}
        onClick={async () => {
          if (resource?.meta?.name?.name)
            await createCanvasDashboardFromMetricsView(
              instanceId,
              resource.meta.name.name,
            );
        }}
      >
        Generate Canvas Dashboard{$ai ? " with AI" : ""}
      </Button>
    {/if}
    <Button
      type={$allowPrimary ? "primary" : "secondary"}
      disabled={!resource}
      onClick={async () => {
        if (resource)
          await createAndPreviewExplore(queryClient, instanceId, resource);
      }}
    >
      Generate Explore Dashboard{$ai ? " with AI" : ""}
    </Button>
  </div>
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <NavigateOrDropdown resources={dashboards} {builder} />
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end">
      <DropdownMenu.Group>
        {#each dashboards as resource (resource?.meta?.name?.name)}
          {@const label =
            resource?.explore?.state?.validSpec?.displayName ||
            resource?.meta?.name?.name}
          {@const filePath = resource?.meta?.filePaths?.[0]}
          {#if label && filePath}
            <DropdownMenu.Item href={`/files/${removeLeadingSlash(filePath)}`}>
              <ExploreIcon color={resourceColorMapping[ResourceKind.Explore]} />
              {label}
            </DropdownMenu.Item>
          {/if}
        {/each}
        <DropdownMenu.Separator />
        {#if $generateCanvas}
          <DropdownMenu.Item
            on:click={async () => {
              if (resource?.meta?.name?.name)
                await createCanvasDashboardFromMetricsView(
                  instanceId,
                  resource.meta.name.name,
                );
            }}
          >
            <Add />
            Generate Canvas Dashboard{$ai ? " with AI" : ""}
          </DropdownMenu.Item>
        {/if}
        <DropdownMenu.Item
          on:click={async () => {
            if (resource)
              await createAndPreviewExplore(queryClient, instanceId, resource);
          }}
        >
          <Add />
          Generate Explore Dashboard{$ai ? " with AI" : ""}
        </DropdownMenu.Item>
      </DropdownMenu.Group>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
