<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import { removeLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useGetExploresForMetricsView } from "../dashboards/selectors";
  import { resourceColorMapping } from "../entity-management/resource-icon-mapping";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { handleEntityCreate } from "../file-explorer/new-files";
  import CreateExploreDashboardButton from "./CreateExploreDashboardButton.svelte";

  export let resource: V1Resource | undefined;

  $: dashboardsQuery = useGetExploresForMetricsView(
    $runtime.instanceId,
    resource?.meta?.name?.name ?? "",
  );

  $: dashboards = $dashboardsQuery.data ?? [];

  async function handleCreateDashboard() {
    const newFilePath = await handleEntityCreate(
      ResourceKind.Explore,
      resource,
    );
    if (newFilePath) {
      await goto(`/files/${newFilePath}`);
    }
  }
</script>

{#if dashboards?.length === 0}
  <CreateExploreDashboardButton metricsViewResource={resource} />
{:else}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger asChild let:builder>
      <Button type="secondary" builders={[builder]}>
        Go to dashboard
        <CaretDownIcon />
      </Button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="end">
      <DropdownMenu.Group>
        {#each dashboards as resource (resource?.meta?.name?.name)}
          {@const label =
            resource?.explore?.state?.validSpec?.title ??
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
        <DropdownMenu.Item on:click={handleCreateDashboard}>
          <Add />
          Create dashboard
        </DropdownMenu.Item>
      </DropdownMenu.Group>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
