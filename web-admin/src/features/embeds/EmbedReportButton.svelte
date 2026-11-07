<script lang="ts">
  import { reportsEnabled } from "@rilldata/web-common/features/feature-flags.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
</script>

<!-- Reports are created from the dashboard's export menus, which have the query to report on. -->
{#if $reportsEnabled}
  <Button
    href="/-/embed/reports"
    type="secondary"
    noStroke
    label={m.report_form_go_to_reports()}
  >
    <ReportIcon className="text-icon-muted" size="14px" />
  </Button>
{/if}
