<script lang="ts">
  import { reportsEnabled } from "@rilldata/web-common/features/feature-flags.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  // The button only links to the reports page, which can't be reached without navigation.
  const navigationEnabled = !!EmbedStore.getInstance()?.navigationEnabled;
</script>

<!-- Reports are created from the dashboard's export menus, which have the query to report on. -->
{#if $reportsEnabled && navigationEnabled}
  <Button
    href="/-/embed/reports"
    type="secondary"
    compact
    label={m.report_form_go_to_reports()}
  >
    <ReportIcon className="hover:text-fg-accent" size="16px" />
  </Button>
{/if}
