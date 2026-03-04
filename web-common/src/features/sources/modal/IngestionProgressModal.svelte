<script lang="ts">
  import { goto } from "$app/navigation";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    sourceIngestionTracker,
    type IngestionState,
  } from "@rilldata/web-common/features/sources/sources-store";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import {
    WandIcon,
    CheckCircle2Icon,
    AlertCircleIcon,
    Loader2Icon,
  } from "lucide-svelte";
  import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import type { V1Resource } from "../../../runtime-client";
  import type { HTTPError } from "../../../runtime-client/fetchWrapper";
  import { extractFileName } from "../../entity-management/file-path-utils";
  import { featureFlags } from "../../feature-flags";
  import {
    createCanvasDashboardFromTableWithAgent,
    useCreateMetricsViewWithCanvasAndExploreUIAction,
  } from "../../metrics-views/ai-generation/generateMetricsView";

  const { ai, developerChat } = featureFlags;
  const ingestionState = sourceIngestionTracker.ingestionState;

  $: state = $ingestionState as IngestionState;
  $: open = state !== null;
  $: filePath = state?.filePath ?? "";
  $: sourceName = extractFileName(filePath);
  $: errorMessage = state?.status === "failed" ? state.error : "";

  $: ({ instanceId } = $runtime);

  let fileArtifact: FileArtifact;
  let sourceQuery: CreateQueryResult<V1Resource, HTTPError>;

  $: if (filePath) {
    fileArtifact = fileArtifacts.getFileArtifact(filePath);
    sourceQuery = fileArtifact.getResource(queryClient, instanceId);
  }
  $: sinkConnector = $sourceQuery?.data?.source?.spec?.sinkConnector;

  $: createDashboardFromTable = filePath
    ? useCreateMetricsViewWithCanvasAndExploreUIAction(
        instanceId,
        sinkConnector as string,
        "",
        "",
        sourceName,
        BehaviourEventMedium.Button,
        MetricsEventSpace.Modal,
      )
    : null;

  function close() {
    sourceIngestionTracker.dismiss();
  }

  async function goToSource() {
    await goto(`/files${filePath}`);
    close();
  }

  async function viewYaml() {
    await goto(`/files${filePath}`);
    close();
  }

  async function generateMetrics() {
    if (createDashboardFromTable === null) return;
    close();

    if ($developerChat) {
      await createCanvasDashboardFromTableWithAgent(
        instanceId,
        sinkConnector as string,
        "",
        "",
        sourceName,
      );
    } else {
      await createDashboardFromTable();
    }
  }
</script>

<AlertDialog.Root {open}>
  <AlertDialog.Content>
    {#if state?.status === "loading"}
      <AlertDialog.Title>
        <div class="flex items-center gap-2">
          <Loader2Icon class="w-5 h-5 text-primary-500 animate-spin" />
          Ingesting data...
        </div>
      </AlertDialog.Title>

      <AlertDialog.Description>
        <p class="font-medium">
          Safe to close this window, we'll notify you when complete.
        </p>
        <p class="mt-2 text-sm text-muted-foreground">
          Processing may take several minutes depending on file size. The upload
          continues in the background, and you'll receive a notification when
          it's complete. You can safely close this modal or cancel at any time.
        </p>
      </AlertDialog.Description>

      <AlertDialog.Footer>
        <AlertDialog.Action asChild let:builder>
          <AlertDialog.Cancel asChild let:builder>
            <Button builders={[builder]} onClick={goToSource} type="secondary">
              View this source
            </Button>
          </AlertDialog.Cancel>

          <Button builders={[builder]} onClick={close} type="primary">
            Close
          </Button>
        </AlertDialog.Action>
      </AlertDialog.Footer>
    {:else if state?.status === "ingested"}
      <AlertDialog.Title>
        <div class="flex items-center gap-2">
          <CheckCircle2Icon class="w-5 h-5 text-green-500" />
          Data imported successfully!
        </div>
      </AlertDialog.Title>

      <AlertDialog.Description>
        <span class="font-mono text-fg-primary break-all">{sourceName}</span>
        has been ingested. What would you like to do next?
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
    {:else if state?.status === "failed"}
      <AlertDialog.Title>
        <div class="flex items-center gap-2">
          <AlertCircleIcon class="w-5 h-5 text-red-500" />
          Data import failed
        </div>
      </AlertDialog.Title>

      <AlertDialog.Description>
        <p class="text-destructive break-all">{errorMessage}</p>
      </AlertDialog.Description>

      <AlertDialog.Footer>
        <AlertDialog.Action asChild let:builder>
          <Button builders={[builder]} onClick={viewYaml} type="secondary">
            View YAML
          </Button>
        </AlertDialog.Action>
      </AlertDialog.Footer>
    {/if}
  </AlertDialog.Content>
</AlertDialog.Root>
