<script lang="ts">
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import { getDashboardFromEmbedRoute } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
  import { page } from "$app/state";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import ScheduledReportDialog from "@rilldata/web-common/features/scheduled-reports/ScheduledReportDialog.svelte";

  let activeResource = $derived(
    getDashboardFromEmbedRoute(page.route.id, page.params),
  );

  const { reports } = featureFlags;
  let enableReports = $derived(
    $reports && !!EmbedStore.getInstance()?.userEmail,
  );

  let open = $state(false);
</script>

{#if enableReports}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="secondary" noStroke>
          <ReportIcon className="text-icon-muted" size="14px" />
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content>
      <DropdownMenu.Item href="/-/embed/reports">
        Go to reports
      </DropdownMenu.Item>
      {#if activeResource?.kind === ResourceKind.Explore}
        <DropdownMenu.Item onclick={() => (open = true)}>
          Create report
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

{#if open && activeResource?.name}
  <ScheduledReportDialog
    bind:open
    props={{
      mode: "create",
      query: { metricsViewAggregationRequest: {} },
      exploreName: activeResource.name,
    }}
  />
{/if}
