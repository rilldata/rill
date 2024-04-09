<script lang="ts">
  import {
    Button,
    IconSpaceFixer,
  } from "@rilldata/web-common/components/button";
  import Forward from "@rilldata/web-common/components/icons/Forward.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { Play } from "lucide-svelte";

  export let filePath: string;
  $: metricsDefName = extractFileName(filePath);
  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);

  const queryClient = useQueryClient();

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath);
  $: yaml = $fileQuery?.data?.blob;
  $: allErrors = fileArtifact.getAllErrors(queryClient, $runtime.instanceId);

  let buttonDisabled = true;
  let buttonStatus;

  const viewDashboard = async () => {
    await behaviourEvent.fireNavigationEvent(
      metricsDefName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Workspace,
      MetricsEventScreenName.MetricsDefinition,
      MetricsEventScreenName.Dashboard,
    );
  };

  const TOOLTIP_CTA = "Fix this error to enable your dashboard.";
  // no content
  $: if (!yaml?.length) {
    buttonDisabled = true;
    buttonStatus = [
      "Your metrics definition is empty. Get started by trying one of the options in the editor.",
    ];
  }
  // content & errors
  else if ($allErrors?.length) {
    buttonDisabled = true;
    buttonStatus = [$allErrors[0].message, TOOLTIP_CTA];
  }
  // preview is available
  else {
    buttonStatus = ["Explore your metrics dashboard"];
    buttonDisabled = false;
  }
</script>

<Tooltip alignment="middle" distance={5} location="right">
  <Button
    disabled={buttonDisabled}
    label="Go to dashboard"
    href={`/dashboard/${metricsDefName}`}
    newTab
    on:click={viewDashboard}
    type="brand"
  >
    <Play size="10px" />
    Preview
  </Button>
  <TooltipContent slot="tooltip-content">
    {#each buttonStatus as status}
      <div>{status}</div>
    {/each}
  </TooltipContent>
</Tooltip>
