<script lang="ts">
  import {
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { RFC5322EmailRegex } from "@rilldata/web-common/components/forms/validation.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let open: boolean;
  export let organization: string;
  export let currentBillingPortalAdmin: string | undefined;

  $: selectedBillingPortalAdmin = currentBillingPortalAdmin ?? "";
  $: trimmedEmail = selectedBillingPortalAdmin.trim();
  $: isValidEmail = RFC5322EmailRegex.test(trimmedEmail);
  $: isDifferentEmail = trimmedEmail !== (currentBillingPortalAdmin ?? "");

  const updateOrg = createAdminServiceUpdateOrganization();

  async function handleAssignAsBillingPortalAdmin() {
    try {
      await $updateOrg.mutateAsync({
        org: organization,
        data: {
          billingPortalAdmin: trimmedEmail,
        },
      });

      eventBus.emit("notification", {
        message: m.billing_portal_admin_assigned({ name: trimmedEmail }),
      });
    } catch (error) {
      console.error("Error assigning user as billing portal admin", error);
      eventBus.emit("notification", {
        message: m.billing_portal_admin_reassign_failed(),
        type: "error",
      });
    }

    open = false;
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content class="w-[520px]" noClose>
    <Dialog.Header>
      <Dialog.Title>{m.billing_change_billing_portal_admin()}</Dialog.Title>

      <Dialog.Description>
        <div class="mt-2 my-1">
          {m.billing_enter_portal_admin_email()}
        </div>
        <Input
          id="billingPortalAdmin"
          bind:value={selectedBillingPortalAdmin}
          placeholder="accounting@example.com"
          errors={trimmedEmail && !isValidEmail
            ? m.users_invalid_email()
            : null}
          alwaysShowError
        />
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer class="mt-3">
      <Button type="secondary" onClick={() => (open = false)}
        >{m.billing_cancel()}</Button
      >
      <Button
        type="primary"
        onClick={handleAssignAsBillingPortalAdmin}
        loading={$updateOrg.isPending}
        disabled={!isValidEmail || !isDifferentEmail}
      >
        {m.billing_assign_as_portal_admin()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
