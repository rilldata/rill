<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { BellIcon } from "lucide-svelte";
  import { getDashboardFromEmbedRoute } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
  import { page } from "$app/state";
  import AlertFormDataWrapper from "@rilldata/web-common/features/alerts/AlertFormDataWrapper.svelte";
  import { DialogContent } from "@rilldata/web-common/components/dialog";
  import GuardedDialog from "@rilldata/web-common/components/dialog/GuardedDialog.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

  let activeResource = $derived(
    getDashboardFromEmbedRoute(page.route.id, page.params),
  );

  const { alerts } = featureFlags;
  let enableAlerts = $derived($alerts && !!EmbedStore.getInstance()?.userEmail);

  let open = $state(false);
</script>

{#if enableAlerts}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="secondary" noStroke>
          <BellIcon class="flex-none text-icon-muted" size="14px" />
        </Button>
      {/snippet}
    </DropdownMenu.Trigger>

    <DropdownMenu.Content>
      <DropdownMenu.Item href="/-/embed/alerts">Go to alerts</DropdownMenu.Item>
      {#if activeResource?.kind === ResourceKind.Explore}
        <DropdownMenu.Item onclick={() => (open = true)}>
          Create alert
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

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
      props={{ mode: "create", exploreName: activeResource.name }}
      {onCancel}
      {onClose}
    />
  </DialogContent>
</GuardedDialog>
