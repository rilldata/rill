<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import PreviewButton from "./GoToDashboardButton.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let filePath: string;
  export let showInspectorToggle = true;

  $: metricsDefName = extractFileName(filePath);

  const onChangeCallback = async (
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) => {
    const newRoute = await handleEntityRename(
      $runtime.instanceId,
      e.currentTarget,
      filePath,
      EntityType.MetricsDefinition,
    );
    if (newRoute) await goto(newRoute);
  };

  function showDeployModal() {
    console.log("Deploying dashboard");
  }
</script>

<WorkspaceHeader
  on:change={onChangeCallback}
  {showInspectorToggle}
  titleInput={metricsDefName}
>
  <div slot="cta" class="flex gap-x-2">
    <Tooltip distance={8}>
      <Button on:click={showDeployModal} type="secondary">Deploy</Button>
      <TooltipContent slot="tooltip-content">
        Deploy this dashboard to Rill Cloud
      </TooltipContent>
    </Tooltip>
    <PreviewButton {filePath} />
  </div>
</WorkspaceHeader>
