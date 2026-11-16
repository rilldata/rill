<script lang="ts">
  import { page } from "$app/state";
  import {
    getEmbedAlertExploreName,
    getEmbedCreateAlertState,
  } from "@rilldata/web-admin/features/embeds/embed-alerts.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { DialogContent } from "@rilldata/web-common/components/dialog";
  import GuardedDialog from "@rilldata/web-common/components/dialog/GuardedDialog.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import AlertFormDataWrapper from "@rilldata/web-common/features/alerts/AlertFormDataWrapper.svelte";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import { alertsEnabled } from "@rilldata/web-common/features/feature-flags.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { BellIcon, BellPlusIcon } from "lucide-svelte";

  const runtimeClient = useRuntimeClient();

  // Without navigation the alerts page can't be reached, so only creating an alert is offered.
  const navigationEnabled = !!EmbedStore.getInstance()?.navigationEnabled;

  let exploreName = $derived(
    getEmbedAlertExploreName(page.route.id, page.params),
  );
  let createAlertState = $derived(
    getEmbedCreateAlertState(runtimeClient, exploreName),
  );
  let { canCreate, isCustomTimeRange } = $derived($createAlertState);

  let open = $state(false);
  // The explore whose alert is being created, captured when the dialog opens.
  let dialogExploreName = $state("");

  function openCreateDialog() {
    dialogExploreName = exploreName;
    open = true;
  }
</script>

{#if $alertsEnabled && navigationEnabled}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="secondary" compact label={m.nav_tab_alerts()}>
          <BellIcon class="flex-none hover:text-fg-accent" size="16px" />
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content align="end">
      <DropdownMenu.Item href="/-/embed/alerts">
        {m.alert_form_go_to_alerts()}
      </DropdownMenu.Item>
      {#if canCreate}
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
{:else if $alertsEnabled && canCreate}
  <Tooltip.Root disabled={!isCustomTimeRange}>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        <Button
          {...props}
          type="secondary"
          compact
          disabled={isCustomTimeRange}
          label={m.alert_create_alert()}
          onClick={openCreateDialog}
        >
          <BellPlusIcon class="flex-none hover:text-fg-accent" size="16px" />
        </Button>
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Content side="bottom">
      {m.alert_set_non_custom_time_range()}
    </Tooltip.Content>
  </Tooltip.Root>
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
