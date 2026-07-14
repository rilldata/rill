<script lang="ts">
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CreatePersonalCanvasDialog from "@rilldata/web-admin/features/personal-files/canvas/CreatePersonalCanvasDialog.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import ExploreIcon from "@rilldata/web-common/components/icons/ExploreIcon.svelte";
  import ResourceList from "@rilldata/web-admin/features/resources/ResourceList.svelte";
  import ResourceListEmptyState from "@rilldata/web-admin/features/resources/ResourceListEmptyState.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { renderComponent } from "tanstack-table-8-svelte-5";
  import PersonalCanvasCompositeCell from "@rilldata/web-admin/features/personal-files/canvas/PersonalCanvasCompositeCell.svelte";
  import { getPersonalFilteredResources } from "@rilldata/web-admin/features/personal-files/selectors.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

  let {
    org,
    project,
  }: {
    org: string;
    project: string;
  } = $props();

  const runtimeClient = useRuntimeClient();
  let personalCanvasesQuery = $derived(
    getPersonalFilteredResources(
      runtimeClient,
      org,
      project,
      ResourceKind.Canvas,
    ),
  );
  let personalCanvases = $derived($personalCanvasesQuery.data ?? []);

  let columns = $derived([
    {
      id: "composite",
      cell: ({ row }) => {
        const resource = row.original as V1Resource;
        const name = resource.meta.name.name;

        // If not a Metrics Explorer, it's a Custom Dashboard.
        const isMetricsExplorer = !!resource?.explore;
        const title = isMetricsExplorer
          ? resource.explore.spec.displayName
          : resource.canvas.spec.displayName;
        const refreshedOn =
          (isMetricsExplorer
            ? resource.explore?.state?.dataRefreshedOn
            : resource.canvas?.state?.dataRefreshedOn) ||
          resource.meta.stateUpdatedOn;

        return renderComponent(PersonalCanvasCompositeCell, {
          name,
          title,
          lastRefreshed: refreshedOn,
          error: resource.meta.reconcileError,
          isMetricsExplorer,
          organization: org,
          project,
        });
      },
    },
  ]);
</script>

{#if personalCanvases.length || $personalCanvasesQuery.isPending}
  <section class="flex flex-col gap-3">
    <header class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <Lock size="16px" />
        <h2 class="text-lg font-medium">My dashboards</h2>
        <span class="text-sm text-fg-secondary">Only visible to you</span>
      </div>
      <CreatePersonalCanvasDialog {org} {project} />
    </header>

    {#if $personalCanvasesQuery.isPending}
      <div class="m-auto mt-20">
        <DelayedSpinner isLoading={true} size="24px" />
      </div>
    {:else}
      <div class="flex flex-col w-full gap-y-3">
        <ResourceList
          kind="personal canvases"
          data={personalCanvases}
          {columns}
          toolbar={false}
        >
          <ResourceListEmptyState
            slot="empty"
            icon={ExploreIcon}
            message="You don't have any personal canvases yet."
          >
            <span slot="action">
              Create one to explore the project's data your way.
            </span>
          </ResourceListEmptyState>
        </ResourceList>
      </div>
    {/if}
  </section>
{/if}
