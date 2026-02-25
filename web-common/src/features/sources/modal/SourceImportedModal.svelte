<script lang="ts">
  import { goto } from "$app/navigation";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { sourceIngestionTracker } from "@rilldata/web-common/features/sources/sources-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import type { V1Resource } from "../../../runtime-client";
  import type { HTTPError } from "../../../runtime-client/fetchWrapper";
  import { extractFileName } from "../../entity-management/file-path-utils";
  import { featureFlags } from "../../feature-flags";
  import {
    useCreateMetricsViewFromTableUIAction,
    useCreateMetricsViewWithCanvasAndExploreUIAction,
  } from "../../metrics-views/ai-generation/generateMetricsView";

  const { ai, generateCanvas } = featureFlags;

  export let sourcePath: string | null;

  const runtimeClient = useRuntimeClient();

  let fileArtifact: FileArtifact;
  let sourceQuery: CreateQueryResult<V1Resource, HTTPError>;

  $: sourceName = extractFileName(sourcePath ?? "");

  $: ({ instanceId } = runtimeClient);

  $: if (sourcePath) {
    fileArtifact = fileArtifacts.getFileArtifact(sourcePath);
    sourceQuery = fileArtifact.getResource(queryClient, instanceId);
  }
  $: sinkConnector = $sourceQuery?.data?.source?.spec?.sinkConnector;

  // When generateCanvas is enabled, create both explore and canvas dashboards
  // and navigate to canvas (with fallback to explore on failure)
  $: createDashboardFromTable =
    sourcePath !== null
      ? $generateCanvas
        ? useCreateMetricsViewWithCanvasAndExploreUIAction(
            runtimeClient,
            instanceId,
            sinkConnector as string,
            "",
            "",
            sourceName,
            BehaviourEventMedium.Button,
            MetricsEventSpace.Modal,
          )
        : useCreateMetricsViewFromTableUIAction(
            runtimeClient,
            instanceId,
            sinkConnector as string,
            "",
            "",
            sourceName,
            true,
            BehaviourEventMedium.Button,
            MetricsEventSpace.Modal,
          )
      : null;

  function close() {
    sourceIngestionTracker.dismiss();
  }

  async function goToSource() {
    await goto(`/files${sourcePath ?? ""}`);
    close();
  }

  async function generateMetrics() {
    // This should never happen, because the button is
    // disabled when this is null, but adding this check
    // for type narrowing and just in case.
    if (createDashboardFromTable === null) return;
    close();
    await createDashboardFromTable();
  }
</script>

<AlertDialog.Root open={sourcePath !== null}>
  <AlertDialog.Content>
    <AlertDialog.Title>Source imported successfully</AlertDialog.Title>

    <AlertDialog.Description>
      <span class="font-mono text-fg-primary break-all">{sourceName}</span> has been
      ingested. What would you like to do next?
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Action asChild let:builder>
        <AlertDialog.Cancel asChild let:builder>
          <Button builders={[builder]} onClick={goToSource} type="secondary">
            View this source
          </Button>
        </AlertDialog.Cancel>

        <Button
          builders={[builder]}
          disabled={createDashboardFromTable === null}
          onClick={generateMetrics}
          type="primary"
        >
          Generate dashboard

          {#if $ai}
            with AI
            <WandIcon class="w-3 h-3" />
          {/if}
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
