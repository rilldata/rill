<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import ResponsiveButtonText from "@rilldata/web-common/components/panel/ResponsiveButtonText.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import EnterIcon from "../../components/icons/EnterIcon.svelte";
  import { WorkspaceHeader } from "../../layout/workspace";
  import { behaviourEvent } from "../../metrics/initMetrics";
  import { BehaviourEventMedium } from "../../metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "../../metrics/service/MetricsTypes";
  import { createModelFromSourceV2 } from "../sources/createModel";

  export let tableName: string;

  const queryClient = useQueryClient();

  const handleCreateModelFromSource = async () => {
    const modelName = await createModelFromSourceV2(queryClient, tableName);
    goto(`/model/${modelName}`);
    behaviourEvent.fireNavigationEvent(
      modelName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.RightPanel,
      MetricsEventScreenName.ExternalTable,
      MetricsEventScreenName.Model,
    );
  };

  function isHeaderWidthSmall(width: number) {
    return width < 800;
  }
</script>

<div class="grid items-center" style:grid-template-columns="auto max-content">
  <WorkspaceHeader
    editable={false}
    onChangeCallback={undefined}
    showInspectorToggle={false}
    {...{ titleInput: tableName }}
  >
    <svelte:fragment slot="cta" let:width={headerWidth}>
      <PanelCTA side="right">
        <Button on:click={handleCreateModelFromSource}>
          <ResponsiveButtonText collapse={isHeaderWidthSmall(headerWidth)}>
            Create model
          </ResponsiveButtonText>
          <IconSpaceFixer pullLeft pullRight={isHeaderWidthSmall(headerWidth)}>
            <EnterIcon size="14px" />
          </IconSpaceFixer>
        </Button>
      </PanelCTA>
    </svelte:fragment>
  </WorkspaceHeader>
</div>
