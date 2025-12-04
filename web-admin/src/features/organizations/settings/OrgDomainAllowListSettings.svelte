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

  export let organization: string;

  $: userDomain = getUserDomain();
  $: isPublicDomain = userDomainIsPublic();

  $: allowedDomains = createAdminServiceListWhitelistedDomains(organization);
  $: domainAllowed = !!$allowedDomains.data?.domains?.find(
    (d) => d.domain === $userDomain.data,
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

<SettingsContainer title="Allow domain access">
  <div slot="body" class="mt-1">
    <div class="flex flex-row items-center gap-x-2">
      {#if !$isPublicDomain.data}
        <Label for="allow-domain" class="font-normal text-gray-700 text-sm">
          Allow existing and new Rill users with a <b>@{$userDomain.data}</b>
          email address to join this org as a <b>Viewer</b>.
          <a
            target="_blank"
            href="https://docs.rilldata.com/reference/cli/user/whitelist"
          >
            Learn more
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
            on:click={updateAllowedDomain}
          />
        </DelayedCircleOutlineSpinner>
      {:else}
        Domain allowlisting is not allowed with a public domain.
        <a
          target="_blank"
          href="https://docs.rilldata.com/reference/cli/user/whitelist"
        >
          Learn more
        </a>
      {/if}
    </div>

    <div class="mt-2 font-medium text-sm">
      <div>Domains added to allowlist by other admins</div>
      {#if $allowedDomains.data?.domains?.length}
        <div class="flex flex-col ml-2 mt-1 gap-y-1">
          {#each $allowedDomains.data.domains as { domain } (domain)}
            <div class="text-gray-500 font-normal">@{domain}</div>
          {/each}
        </div>
      {:else}
        <div class="text-gray-500">none</div>
      {/if}
    </div>
  </div>
</SettingsContainer>
