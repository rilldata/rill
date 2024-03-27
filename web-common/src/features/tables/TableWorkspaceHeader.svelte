<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import Add from "../../components/icons/Add.svelte";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { BehaviourEventMedium } from "../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../metrics/service/MetricsTypes";
  import { runtime } from "../../runtime-client/runtime-store";
  import { useCreateDashboardFromTableUIAction } from "../metrics-views/ai-generation/generateMetricsView";

  export let fullyQualifiedTableName: string;

  $: tableName = fullyQualifiedTableName.split(".")[1];

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    tableName,
    BehaviourEventMedium.Button,
    MetricsEventSpace.RightPanel,
  );

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    editable={false}
    showInspectorToggle={false}
    {...{ titleInput: fullyQualifiedTableName }}
  >
    <svelte:fragment slot="cta" let:width={headerWidth}>
      {@const collapse = isHeaderWidthSmall(headerWidth)}
      <PanelCTA side="right">
        <Button on:click={createDashboardFromTable}>
          <IconSpaceFixer pullLeft pullRight={collapse}>
            <Add />
          </IconSpaceFixer>
          <ResponsiveButtonText {collapse}>
            Generate dashboard with AI
          </ResponsiveButtonText>
        </Button>
      </PanelCTA>
    </svelte:fragment>
  </WorkspaceHeader>
</div>
