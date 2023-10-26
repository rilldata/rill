<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createDownloadReportMutation,
    DownloadFormatMap,
  } from "@rilldata/web-admin/features/projects/report-download";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: reportId = $page.params.report;
  $: format = $page.url.searchParams.get("format");
  $: bakedQuery = $page.url.searchParams.get("query");

  const downloadReportMutation = createDownloadReportMutation();
  let once = false;

  $: if (reportId && format && bakedQuery && $runtime) {
    if (!once) {
      once = true;
      $downloadReportMutation.mutateAsync({
        data: {
          instanceId: $runtime.instanceId,
          format,
          bakedQuery,
        },
      });
      // TODO: redirect to report page once success once that is merged in
    }
  }

  let error: string;
  $: if (!(format in DownloadFormatMap)) {
    error = !format ? "format required" : `unknown format: "${format}"`;
  } else if (!bakedQuery) {
    error = "query required";
  } else if ($downloadReportMutation.error) {
    error = $downloadReportMutation.error.response.data.message;
  }
</script>

{#if error}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <CtaMessage>
        {error}
      </CtaMessage>
      <!-- TODO: redirect to report once that page is merged -->
      <CtaButton variant="primary-outline" on:click={() => goto("/")}>
        Back to home
      </CtaButton>
    </CtaContentContainer>
  </CtaLayoutContainer>
{/if}
