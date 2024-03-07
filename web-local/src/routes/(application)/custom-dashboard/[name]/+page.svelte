<script lang="ts">
  import { page } from "$app/stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useCustomDashboard } from "@rilldata/web-common/features/custom-dashboards/selectors";
  import CustomDashboard from "@rilldata/web-common/features/custom-dashboards/CustomDashboard.svelte";
  import CustomDashboardEditor from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEditor.svelte";
  import {
    WorkspaceContainer,
    WorkspaceHeader,
  } from "@rilldata/web-common/layout/workspace";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { goto } from "$app/navigation";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { getRouteFromName } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { isDuplicateName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";

  let editing = true;
  let showGrid = false;

  $: customDashboardName = $page.params.name;

  $: query = useCustomDashboard($runtime.instanceId, customDashboardName);

  $: dashboard = $query.data?.dashboard?.spec;

  $: columns = Number(dashboard?.gridColumns ?? 12);
  $: gap = Number(dashboard?.gridGap ?? 1);

  $: allNamesQuery = useAllNames($runtime.instanceId);

  $: charts = dashboard?.components ?? ([] as V1DashboardComponent[]);

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    if (!e.currentTarget) return;
    if (!e.currentTarget.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Model name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.currentTarget.value = customDashboardName; // resets the input
      return;
    }
    if (
      isDuplicateName(
        e.currentTarget.value,
        customDashboardName,
        $allNamesQuery?.data ?? [],
      )
    ) {
      notifications.send({
        message: `Name ${e.currentTarget.value} is already in use`,
      });
      e.currentTarget.value = customDashboardName; // resets the input
      return;
    }

    try {
      const toName = e.currentTarget.value;
      const entityType = EntityType.Dashboard;
      await renameFileArtifact(
        $runtime.instanceId,
        customDashboardName,
        toName,
        entityType,
      );
      await goto(getRouteFromName(toName, entityType), {
        replaceState: true,
      });
    } catch (err) {
      console.error(err.response.data.message);
    }
  };
</script>

<svelte:head>
  <title>Rill Developer | {customDashboardName}</title>
</svelte:head>

<WorkspaceContainer assetID={customDashboardName} inspector={false}>
  <WorkspaceHeader
    slot="header"
    titleInput={customDashboardName}
    showInspectorToggle={false}
    {onChangeCallback}
  >
    <div slot="workspace-controls" class="flex gap-x-4">
      {#if !editing}
        <Button
          on:click={() => {
            showGrid = !showGrid;
          }}
        >
          {showGrid ? "Show grid" : "Hide grid"}
        </Button>
      {/if}
      <Button
        on:click={async () => {
          if (editing) {
            await $query.refetch();
          }
          editing = !editing;
        }}
      >
        {editing ? "Preview" : "Edit"}
      </Button>
    </div>
  </WorkspaceHeader>

  <svelte:fragment slot="body">
    {#if editing}
      <CustomDashboardEditor {customDashboardName} />
    {:else}
      <CustomDashboard {showGrid} {columns} {charts} {gap} />
    {/if}
  </svelte:fragment>
</WorkspaceContainer>
