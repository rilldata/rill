<script lang="ts">
  import {
    createAdminServiceCreateWhitelistedDomain,
    createAdminServiceListWhitelistedDomains,
    createAdminServiceRemoveWhitelistedDomain,
    getAdminServiceListWhitelistedDomainsQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import {
    getUserDomain,
    userDomainIsPublic,
  } from "@rilldata/web-admin/features/projects/user-management/selectors";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import DelayedCircleOutlineSpinner from "@rilldata/web-common/components/spinner/DelayedCircleOutlineSpinner.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let { organization }: { organization: string } = $props();

  let userDomain = $derived(getUserDomain());
  let isPublicDomain = $derived(userDomainIsPublic());

  let allowedDomains = $derived(
    createAdminServiceListWhitelistedDomains(organization),
  );
  let domainAllowed = $derived(
    !!$allowedDomains.data?.domains?.find((d) => d.domain === $userDomain.data),
  );

  const allowDomainMutation = createAdminServiceCreateWhitelistedDomain();
  const disallowDomainMutation = createAdminServiceRemoveWhitelistedDomain();
  async function updateAllowedDomain() {
    if (domainAllowed) {
      await $disallowDomainMutation.mutateAsync({
        org: organization,
        domain: $userDomain.data,
      });
    } else {
      await $allowDomainMutation.mutateAsync({
        org: organization,
        data: {
          domain: $userDomain.data,
          role: OrgUserRoles.Viewer,
        },
      });
    }

    void queryClient.refetchQueries({
      queryKey: getAdminServiceListWhitelistedDomainsQueryKey(organization),
    });
  }
</script>

<SettingsContainer title={m.settings_allow_domain_title()}>
  <div class="mt-1">
    <div class="flex flex-row items-center gap-x-2">
      {#if !$isPublicDomain.data}
        <Label for="allow-domain" class="font-normal text-fg-secondary text-sm">
          {@html m.settings_allow_domain_description({
            domain: `<b>@${$userDomain.data}</b>`,
            role: `<b>Viewer</b>`,
          })}
          <a
            target="_blank"
            href="https://docs.rilldata.com/reference/cli/user/whitelist"
          >
            {m.settings_learn_more()}
          </a>
        </Label>
        <div class="grow"></div>
        <DelayedCircleOutlineSpinner
          isLoading={$disallowDomainMutation.isPending ||
            $allowDomainMutation.isPending}
        >
          <Switch
            checked={domainAllowed}
            id="allow-domain"
            onclick={updateAllowedDomain}
          />
        </DelayedCircleOutlineSpinner>
      {:else}
        {m.settings_domain_not_allowed_public()}
        <a
          target="_blank"
          href="https://docs.rilldata.com/reference/cli/user/whitelist"
        >
          {m.settings_learn_more()}
        </a>
      {/if}
    </div>

    <div class="mt-2 font-medium text-sm">
      <div>{m.settings_domains_added_by_admins()}</div>
      {#if $allowedDomains.data?.domains?.length}
        <div class="flex flex-col ml-2 mt-1 gap-y-1">
          {#each $allowedDomains.data.domains as { domain } (domain)}
            <div class="text-fg-secondary font-normal">@{domain}</div>
          {/each}
        </div>
      {:else}
        <div class="text-fg-secondary">{m.settings_none()}</div>
      {/if}
    </div>
  </div>
</SettingsContainer>
