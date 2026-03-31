<script lang="ts">
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import { Button } from "../../components/button";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import { featureFlags } from "../feature-flags";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import ViewAsButton from "../dashboards/granular-access-policies/ViewAsButton.svelte";
  import {
    useDashboardPolicyCheck,
    useRillYamlPolicyCheck,
  } from "../dashboards/granular-access-policies/useSecurityPolicyCheck";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";

  const client = useRuntimeClient();

  export let canvasName: string;

  $: canvasQuery = useCanvas(client, canvasName);
  $: canvasFilePath = $canvasQuery.data?.filePath ?? "";

  $: canvasPolicyCheck = useDashboardPolicyCheck(client, canvasFilePath);
  $: rillYamlPolicyCheck = useRillYamlPolicyCheck(client);

  // Check if any metrics view in the project has security rules
  $: metricsViewResources = useFilteredResources(
    client,
    ResourceKind.MetricsView,
    (data) => {
      return (
        data?.resources?.some(
          (res) =>
            (res.metricsView?.state?.validSpec?.securityRules?.length ?? 0) > 0,
        ) ?? false
      );
    },
  );

  $: hasSecurityPolicy =
    $canvasPolicyCheck.data ||
    $rillYamlPolicyCheck.data ||
    $metricsViewResources.data;

  const { dashboardChat, readOnly } = featureFlags;
</script>

<div class="flex gap-2 flex-shrink-0 ml-auto">
  {#if hasSecurityPolicy}
    <ViewAsButton />
  {/if}
  {#if $dashboardChat}
    <ChatToggle />
  {/if}
  {#if !$readOnly}
    <Button type="secondary" href={`/files${canvasFilePath}`}>Edit</Button>
  {/if}
</div>
