<script lang="ts">
  import { page } from "$app/stores";
  import { createDownloadReportMutation } from "@rilldata/web-admin/features/projects/download-report";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: reportId = $page.params.report;
  $: format = $page.url.searchParams.get("format");
  $: limit = $page.url.searchParams.get("limit");
  $: executionTime = $page.url.searchParams.get("execution_time");
  $: token = $page.url.searchParams.get("token");

  const downloadReportMutation = createDownloadReportMutation();
  let downloadOnce = false;

  function triggerDownload() {
    if (downloadOnce) return;
    downloadOnce = true;
    $downloadReportMutation.mutateAsync({
      data: {
        instanceId: $runtime.instanceId,
        reportId,
        format: (format as V1ExportFormat) ?? V1ExportFormat.EXPORT_FORMAT_CSV,
        executionTime,
        limit,
      },
    });
  }

  $: if (reportId && $runtime) {
    triggerDownload();
  }

  let error: string;
  $: if ($downloadReportMutation.error) {
    error =
      $downloadReportMutation.error.response?.data?.message ?? "unknown error";
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Download failed</h2>
        <CtaMessage>
          {error}
        </CtaMessage>
      </div>
    {:else}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Downloading report...</h2>
        <CtaMessage>
          If your download fails, refresh the page to try again.
        </CtaMessage>
      </div>
    {/if}
    <!-- User accessing with token wont have access to view report. So only show for other rows. -->
    {#if !token}
      <CtaButton
        variant="secondary"
        href={`/${organization}/${project}/-/reports/${reportId}`}
      >
        Go to report page
      </CtaButton>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
