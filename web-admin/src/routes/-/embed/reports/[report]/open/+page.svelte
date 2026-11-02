<script lang="ts">
  import ReportOpen from "@rilldata/web-admin/features/scheduled-reports/open/ReportOpen.svelte";
  import { EmbedDashboardPaths } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  let backHref = $derived(`/-/embed/reports/${data.reportId}`);
</script>

{#if data.isAIReport}
  <CtaLayoutContainer>
    <CtaContentContainer>
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">{m.report_unable_to_open()}</h2>
        <CtaMessage>{m.report_ai_open_not_supported_in_embed()}</CtaMessage>
      </div>
      <CtaButton href={backHref} variant="secondary">
        {m.report_go_to_page()}
      </CtaButton>
    </CtaContentContainer>
  </CtaLayoutContainer>
{:else}
  <!-- Embeds have no magic token, so the report's filters must be applied by the dashboard. -->
  <ReportOpen
    reportSpec={data.reportSpec}
    exploreName={data.exploreName}
    executionTime={data.executionTime}
    dashboardPaths={EmbedDashboardPaths}
    ignoreFilters={false}
    {backHref}
  />
{/if}
