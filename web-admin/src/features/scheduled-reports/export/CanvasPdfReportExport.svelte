<script lang="ts">
  import { page } from "$app/stores";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/LoadingSpinner.svelte";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { getCanvasStoreUnguarded } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { exportCanvasPdf } from "@rilldata/web-common/features/exports/pdf/export-canvas-pdf";
  import type { ExportProgress } from "@rilldata/web-common/features/exports/pdf/types";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let canvasName: string;
  export let includeFilters = true;
  export let allTabs = true;

  const runtimeClient = useRuntimeClient();
  // The export can take a while for large canvases since every component query must complete.
  const EXPORT_TIMEOUT_MS = 120_000;

  $: ({ instanceId } = runtimeClient);
  $: ({ organization, project, report } = $page.params);
  $: token = $page.url.searchParams.get("token");
  $: executionTime = $page.url.searchParams.get("execution_time");

  let state: "generating" | "done" | "error" = "generating";
  let errorMessage = "";
  let progress: ExportProgress | undefined = undefined;
  let started = false;

  // The PDF footer links to the canvas dashboard (with the report's state applied),
  // not to this export page: the export page URL would re-trigger a download.
  function buildDashboardUrl(): string {
    const dashboardUrl = new URL(
      token
        ? `/${organization}/${project}/-/share/${token}/canvas/${canvasName}`
        : `/${organization}/${project}/canvas/${canvasName}`,
      window.location.origin,
    );
    const stateParams = new URLSearchParams(window.location.search);
    stateParams.delete("token");
    stateParams.delete("execution_time");
    stateParams.forEach((value, key) =>
      dashboardUrl.searchParams.set(key, value),
    );
    return dashboardUrl.toString();
  }

  async function runExport() {
    if (started) return;
    started = true;
    state = "generating";
    try {
      await exportCanvasPdf({
        canvasName,
        instanceId,
        includeFilters,
        allTabs,
        timeoutMs: EXPORT_TIMEOUT_MS,
        dashboardUrl: buildDashboardUrl(),
        onProgress: (p) => (progress = p),
      });
      state = "done";
    } catch (e) {
      errorMessage = extractErrorMessage(e);
      state = "error";
    }
  }

  function downloadAgain() {
    started = false;
    void runExport();
  }

  // Runs when the canvas provider's slot content mounts, i.e. once the canvas store is initialized.
  function startExportOnMount(_node: HTMLElement) {
    // Anchor relative time ranges at the report's execution time,
    // so the PDF shows data for the scheduled run rather than the download time.
    if (executionTime) {
      getCanvasStoreUnguarded(
        canvasName,
        instanceId,
      )?.canvasEntity.timeManager.executionTimeStore.set(executionTime);
    }
    void runExport();
  }

  const PROGRESS_COPY: Record<ExportProgress["phase"], () => string> = {
    preparing: () => m.export_pdf_rendering_charts(),
    capturing: () => m.export_pdf_capturing(),
    assembling: () => m.export_pdf_building(),
  };
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if state === "error"}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">{m.report_pdf_export_failed()}</h2>
        <CtaMessage>
          {errorMessage}
        </CtaMessage>
      </div>
      <CtaButton variant="secondary" onClick={downloadAgain}>
        {m.report_pdf_export_retry()}
      </CtaButton>
    {:else if state === "done"}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">{m.report_pdf_export_done()}</h2>
        <CtaMessage>
          {m.report_pdf_export_done_hint()}
        </CtaMessage>
      </div>
      <CtaButton variant="secondary" onClick={downloadAgain}>
        {m.report_pdf_export_download_again()}
      </CtaButton>
    {:else}
      <div class="mt-10">
        <LoadingSpinner />
      </div>
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">{m.report_pdf_export_preparing()}</h2>
        <CtaMessage>
          {progress ? PROGRESS_COPY[progress.phase]() : ""}
        </CtaMessage>
      </div>
    {/if}
    <!-- User accessing with token wont have access to view report. So only show for other rows. -->
    {#if !token}
      <CtaButton
        variant="secondary"
        href={`/${organization}/${project}/-/reports/${report}`}
      >
        {m.report_go_to_page()}
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>

<!-- Off-screen but rendered canvas: the PDF pipeline's capture target must be laid out
     and painted, so the host is positioned off-viewport instead of display:none. -->
<div class="canvas-host" aria-hidden="true">
  {#key `${instanceId}::${canvasName}`}
    <CanvasProvider {canvasName} {instanceId}>
      <div class="size-full" use:startExportOnMount>
        <CanvasDashboardEmbed {canvasName} navigationEnabled={false} />
      </div>
    </CanvasProvider>
  {/key}
</div>

<style lang="postcss">
  .canvas-host {
    position: fixed;
    top: 0;
    left: -200vw;
    width: 1440px;
    height: 900px;
    overflow: hidden;
    pointer-events: none;
  }
</style>
