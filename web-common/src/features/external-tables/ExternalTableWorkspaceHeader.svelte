<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Add from "../../components/icons/Add.svelte";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../metrics/service/MetricsTypes";
  import { createDashboardFromExternalTable } from "./createDashboardFromExternalTable";

  export let fullyQualifiedTableName: string;

  $: tableName = fullyQualifiedTableName.split(".")[1];

  const queryClient = useQueryClient();

  async function handleCreateDashboardFromExternalTable() {
    const newDashboardName = await createDashboardFromExternalTable(
      queryClient,
      tableName,
    );
    goto(`/dashboard/${newDashboardName}`);
    behaviourEvent.fireNavigationEvent(
      newDashboardName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.ExternalTable,
      MetricsEventScreenName.Dashboard,
    );
  }

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    editable={false}
    onChangeCallback={undefined}
    showInspectorToggle={false}
    {...{ titleInput: fullyQualifiedTableName }}
  >
    <svelte:fragment slot="cta" let:width={headerWidth}>
      {@const collapse = isHeaderWidthSmall(headerWidth)}
      <PanelCTA side="right">
        <Button on:click={handleCreateDashboardFromExternalTable}>
          <IconSpaceFixer pullLeft pullRight={collapse}>
            <Add />
          </IconSpaceFixer>
          <ResponsiveButtonText {collapse}>
            Autogenerate dashboard
          </ResponsiveButtonText>
        </Button>
      </PanelCTA>
    </svelte:fragment>
  </WorkspaceHeader>
</div>
