<script lang="ts">
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import { Button } from "../../components/button";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import { featureFlags } from "../feature-flags";
  import { getFileHref } from "../workspaces/edit-routing";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import ViewAsButton from "../dashboards/granular-access-policies/ViewAsButton.svelte";
  import {
    useDashboardPolicyCheck,
    useRillYamlPolicyCheck,
  } from "../dashboards/granular-access-policies/useSecurityPolicyCheck";

  const client = useRuntimeClient();

  export let canvasName: string;

  $: canvasQuery = useCanvas(client, canvasName);
  $: canvasFilePath = $canvasQuery.data?.filePath ?? "";

  $: canvasPolicyCheck = useDashboardPolicyCheck(client, canvasFilePath);
  $: rillYamlPolicyCheck = useRillYamlPolicyCheck(client);

  // Check if any metrics view referenced by this canvas has security rules
  $: referencedMetricsViewsHavePolicy = Object.values(
    $canvasQuery.data?.metricsViews ?? {},
  ).some((mv) => (mv?.state?.validSpec?.securityRules?.length ?? 0) > 0);

  $: hasSecurityPolicy =
    $canvasPolicyCheck.data ||
    $rillYamlPolicyCheck.data ||
    referencedMetricsViewsHavePolicy;

  const { dashboardChat, readOnly } = featureFlags;

  $: hasAnyContent = hasSecurityPolicy || $dashboardChat || !$readOnly;
</script>

{#if hasAnyContent}
  <div class="flex gap-2 flex-shrink-0 ml-auto">
    {#if hasSecurityPolicy}
      <ViewAsButton />
    {/if}
    {#if $dashboardChat}
      <ChatToggle />
    {/if}
    {#if !$readOnly}
      <Button type="secondary" href={getFileHref(canvasFilePath)}>Edit</Button>
    {/if}
  </div>
{/if}
