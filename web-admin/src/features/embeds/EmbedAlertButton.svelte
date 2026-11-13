<script lang="ts">
  import { page } from "$app/state";
  import { getDashboardFromEmbedRoute } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { DialogContent } from "@rilldata/web-common/components/dialog";
  import GuardedDialog from "@rilldata/web-common/components/dialog/GuardedDialog.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import AlertFormDataWrapper from "@rilldata/web-common/features/alerts/AlertFormDataWrapper.svelte";
  import { useExploreState } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
  import { alertsEnabled } from "@rilldata/web-common/features/feature-flags.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import {
    TimeComparisonOption,
    TimeRangePreset,
  } from "@rilldata/web-common/lib/time/types";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { BellIcon } from "lucide-svelte";

  const runtimeClient = useRuntimeClient();

  // Alerts are created on an explore, so they can only be created from explore pages.
  let exploreName = $derived.by(() => {
    const dashboard = getDashboardFromEmbedRoute(page.route.id, page.params);
    return dashboard?.kind === ResourceKind.Explore ? dashboard.name : "";
  });

  // The header is outside of the explore's state managers, so the explore's spec and state are read by name.
  let exploreSpecQuery = $derived(
    useExploreValidSpec(runtimeClient, exploreName),
  );
  let hasTimeDimension = $derived(
    !!$exploreSpecQuery.data?.metricsView?.timeDimension,
  );
  let exploreState = $derived(useExploreState(exploreName));
  // Same as the guard of the dashboard's create alert button (see CreateAlert).
  let isCustomTimeRange = $derived(
    $exploreState?.selectedTimeRange?.name === TimeRangePreset.CUSTOM ||
      $exploreState?.selectedComparisonTimeRange?.name ===
        TimeComparisonOption.CUSTOM,
  );

  let open = $state(false);
  // The explore whose alert is being created, captured when the dialog opens.
  let dialogExploreName = $state("");

  function openCreateDialog() {
    dialogExploreName = exploreName;
    open = true;
  }
</script>

{#if $alertsEnabled}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="secondary" noStroke label={m.nav_tab_alerts()}>
          <BellIcon class="flex-none hover:text-fg-accent" size="16px" />
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content align="end">
      <DropdownMenu.Item href="/-/embed/alerts">
        {m.alert_form_go_to_alerts()}
      </DropdownMenu.Item>
      {#if exploreName && hasTimeDimension && $exploreState}
        <DropdownMenu.Item
          onclick={openCreateDialog}
          disabled={isCustomTimeRange}
        >
          {m.alert_create_alert()}
        </DropdownMenu.Item>
        {#if isCustomTimeRange}
          <div class="px-2 py-1 text-xs text-fg-secondary max-w-[220px]">
            {m.alert_set_non_custom_time_range()}
          </div>
        {/if}
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

{#if dialogExploreName}
  <GuardedDialog
    title={m.dialog_close_without_saving_title()}
    description={m.dialog_close_without_saving_alert_desc()}
    confirmLabel={m.dialog_close_without_saving_confirm()}
    cancelLabel={m.dialog_close_without_saving_cancel()}
    bind:open
    let:onCancel
    let:onClose
    let:preventClose
  >
    <DialogContent
      class="p-0 m-0 max-w-[min(802px,calc(100vw-2rem))] rounded-md"
      noClose
      onEscapeKeydown={preventClose}
      onInteractOutside={preventClose}
    >
      <AlertFormDataWrapper
        props={{ mode: "create", exploreName: dialogExploreName }}
        {onCancel}
        {onClose}
      />
    </DialogContent>
  </GuardedDialog>
{/if}
