<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { createDownloadReportMutation } from "@rilldata/web-admin/features/projects/download-report";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import type { V1ExportFormat } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: reportId = $page.params.report;
  $: format = $page.url.searchParams.get("format");
  $: bakedQuery = $page.url.searchParams.get("query");
  $: limit = $page.url.searchParams.get("limit");

  const downloadReportMutation = createDownloadReportMutation();
  let downloadOnce = false;

  function triggerDownload() {
    if (downloadOnce) return;
    downloadOnce = true;
    $downloadReportMutation.mutateAsync({
      data: {
        instanceId: $runtime.instanceId,
        format: format as V1ExportFormat,
        bakedQuery,
        limit,
      },
    });
  }

  $: if (reportId && format && bakedQuery && $runtime) {
    triggerDownload();
  }

  let error: string;
  $: {
    if (!format) {
      error = "format is required";
    } else if (!bakedQuery) {
      error = "query is required";
    } else if ($downloadReportMutation.error) {
      error =
        $downloadReportMutation.error.response?.data?.message ??
        "unknown error";
    }
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
    <CtaButton
      on:click={() => goto(`/${organization}/${project}/-/reports/${reportId}`)}
      variant="primary-outline"
    >
      Go to report page
    </CtaButton>
  </CtaContentContainer>
</CtaLayoutContainer>
