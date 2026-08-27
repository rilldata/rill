<script lang="ts">
  import { useExplore } from "@rilldata/web-common/features/explores/selectors.ts";
  import {
    useDashboardPolicyCheck,
    useRillYamlPolicyCheck,
  } from "@rilldata/web-common/features/dashboards/granular-access-policies/useSecurityPolicyCheck.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector.ts";
  import ViewAsButton from "@rilldata/web-common/features/dashboards/granular-access-policies/ViewAsButton.svelte";
  import { page } from "$app/state";
  import { createUpdateEditSessionDevJWT } from "@rilldata/web-admin/features/edit-session/updateEditSessionDevJWT.ts";
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";

  let { organization, project }: { organization: string; project: string } =
    $props();

  const runtimeClient = useRuntimeClient();

  let branch = extractBranchFromPath(page.url.pathname);
  let projectQuery = $derived(
    createAdminServiceGetProject(
      organization,
      project,
      branch ? { branch } : undefined,
    ),
  );
  let deploymentId = $derived($projectQuery.data?.deployment?.id);

  let onExplorePreview = $derived(
    !!page.route.id?.startsWith(
      "/[organization]/[project]/-/edit/(viz)/explore",
    ),
  );
  let onCanvasPreview = $derived(
    !!page.route.id?.startsWith(
      "/[organization]/[project]/-/edit/(viz)/canvas",
    ),
  );
  let dashboardName = $derived(page.params.name ?? "");

  // If explore, check explore yaml or its metrics view has security rules
  let exploreQuery = $derived(
    useExplore(runtimeClient, dashboardName, { enabled: onExplorePreview }),
  );
  let exploreFilePath = $derived(
    $exploreQuery.data?.explore?.meta?.filePaths?.[0] ?? "",
  );
  let metricsViewFilePath = $derived(
    $exploreQuery.data?.metricsView?.meta?.filePaths?.[0] ?? "",
  );
  let explorePolicyCheck = $derived(
    useDashboardPolicyCheck(runtimeClient, exploreFilePath),
  );
  let metricsPolicyCheck = $derived(
    useDashboardPolicyCheck(runtimeClient, metricsViewFilePath),
  );

  // Else if canvas, check canvas yaml or its metrics views have security rules
  let canvasQuery = $derived(
    useCanvas(runtimeClient, dashboardName, { enabled: onCanvasPreview }),
  );
  let canvasFilePath = $derived($canvasQuery.data?.filePath ?? "");
  let canvasPolicyCheck = $derived(
    useDashboardPolicyCheck(runtimeClient, canvasFilePath),
  );
  // Check if any metrics view referenced by this canvas has security rules
  let referencedMetricsViewsHavePolicy = $derived(
    Object.values($canvasQuery.data?.metricsViews ?? {}).some(
      (mv) => (mv?.state?.validSpec?.securityRules?.length ?? 0) > 0,
    ),
  );

  // Check rill yaml for direct policy
  let rillYamlPolicyCheck = $derived(useRillYamlPolicyCheck(runtimeClient));

  let hasSecurityPolicy = $derived(
    $explorePolicyCheck.data ||
      $metricsPolicyCheck.data ||
      $canvasPolicyCheck.data ||
      $rillYamlPolicyCheck.data ||
      referencedMetricsViewsHavePolicy,
  );

  let devJTWUpdater = $derived(createUpdateEditSessionDevJWT(deploymentId));
</script>

{#if hasSecurityPolicy}
  <ViewAsButton {devJTWUpdater} />
{/if}
