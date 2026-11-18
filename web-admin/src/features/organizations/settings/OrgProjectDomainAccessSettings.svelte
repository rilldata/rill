<script lang="ts">
  import {
    createAdminServiceGetOrganization,
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import DelayedCircleOutlineSpinner from "@rilldata/web-common/components/spinner/DelayedCircleOutlineSpinner.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import type { AxiosError } from "axios";

  let { organization }: { organization: string } = $props();

  let orgResp = $derived(createAdminServiceGetOrganization(organization));
  let allowed = $derived(
    !$orgResp.data?.organization?.disableProjectDomainAccess,
  );

  const updateOrgMutation = createAdminServiceUpdateOrganization();

  async function toggle() {
    try {
      await $updateOrgMutation.mutateAsync({
        org: organization,
        data: { disableProjectDomainAccess: allowed },
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetOrganizationQueryKey(organization),
      });
    } catch (e) {
      eventBus.emit("notification", {
        type: "error",
        message:
          (e as AxiosError<RpcStatus>).response?.data?.message ??
          (e as Error).message,
      });
    }
  }
</script>

<SettingsContainer title={m.settings_project_domain_access_title()}>
  <div class="mt-1 flex flex-row items-center gap-x-2">
    <Label
      for="allow-project-domain-access"
      class="font-normal text-fg-secondary text-sm"
    >
      {m.settings_project_domain_access_description()}
      <a
        target="_blank"
        rel="noopener noreferrer"
        href="https://docs.rilldata.com/guide/administration/users-and-access/user-management#project-domain-access"
      >
        {m.settings_learn_more()}
      </a>
    </Label>
    <div class="grow"></div>
    <DelayedCircleOutlineSpinner isLoading={$updateOrgMutation.isPending}>
      <Switch
        checked={allowed}
        disabled={!$orgResp.data}
        id="allow-project-domain-access"
        onclick={toggle}
      />
    </DelayedCircleOutlineSpinner>
  </div>
</SettingsContainer>
