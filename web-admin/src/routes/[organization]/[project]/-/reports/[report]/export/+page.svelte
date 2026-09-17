<script lang="ts">
  import { page } from "$app/stores";
  import CanvasPdfReportExport from "@rilldata/web-admin/features/scheduled-reports/export/CanvasPdfReportExport.svelte";
  import { createDownloadReportMutation } from "@rilldata/web-admin/features/projects/download-report";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PageData } from "./$types";

  export let data: PageData;

  const runtimeClient = useRuntimeClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: reportId = $page.params.report;
  $: executionTime = $page.url.searchParams.get("execution_time");
  $: token = $page.url.searchParams.get("token");

  const downloadReportMutation = createDownloadReportMutation(runtimeClient);
  let downloadOnce = false;

  async function triggerDownload() {
    if (downloadOnce) return;
    downloadOnce = true;
    await $downloadReportMutation.mutateAsync({
      data: {
        reportId,
        executionTime,
        originBaseUrl: window.location.origin,
        host: runtimeClient.host,
      },
    });
  }

  $: if (reportId && runtimeClient && !data.canvasName) {
    triggerDownload();
  }

  let error: string;
  $: if ($downloadReportMutation.error) {
    error =
      $downloadReportMutation.error.response?.data?.message ?? "unknown error";
  }
</script>

{#if data.canvasName}
  <!-- Canvas PDF report: render the canvas in the browser and generate the PDF client-side. -->
  <CanvasPdfReportExport
    canvasName={data.canvasName}
    includeFilters={data.pdfIncludeFilters ?? true}
    allTabs={data.pdfAllTabs ?? true}
  />
{:else}
  <CtaLayoutContainer>
    <CtaContentContainer>
      {#if error}
        <div class="flex flex-col gap-y-2">
          <h2 class="text-lg font-semibold">{m.report_download_failed()}</h2>
          <CtaMessage>
            {error}
          </CtaMessage>
        </div>
      {:else}
        <div class="flex flex-col gap-y-2">
          <h2 class="text-lg font-semibold">{m.report_downloading()}</h2>
          <CtaMessage>
            {m.report_download_retry_hint()}
          </CtaMessage>
        </div>
      {/if}
      <!-- User accessing with token wont have access to view report. So only show for other rows. -->
      {#if !token}
        <CtaButton
          variant="secondary"
          href={`/${organization}/${project}/-/reports/${reportId}`}
        >
          {m.report_go_to_page()}
        </CtaButton>
      {/if}
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
