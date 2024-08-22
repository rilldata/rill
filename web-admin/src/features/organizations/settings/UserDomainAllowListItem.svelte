<script lang="ts">
  import {
    createAdminServiceCreateWhitelistedDomain,
    createAdminServiceListWhitelistedDomains,
    createAdminServiceRemoveWhitelistedDomain,
    getAdminServiceListWhitelistedDomainsQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsItemContainer from "@rilldata/web-admin/features/organizations/settings/SettingsItemContainer.svelte";
  import {
    getUserDomain,
    userDomainIsPublic,
  } from "@rilldata/web-admin/features/projects/user-invite/selectors";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
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
        organization,
        domain: $userDomain.data,
      });
    } else {
      await $allowDomainMutation.mutateAsync({
        organization,
        data: {
          domain: $userDomain.data,
          role: "viewer",
        },
      });
    }

    void queryClient.refetchQueries(
      getAdminServiceListWhitelistedDomainsQueryKey(organization),
    );
  }
</script>

<!-- hide if user's domain is not public and no domains are added -->
{#if $isPublicDomain.data || $allowedDomains.data?.domains?.length}
  <SettingsItemContainer title="Allow domain access">
    <div slot="description">
      {#if $isPublicDomain.data}
        <div class="flex flex-col gap-y-1">
          <div class="flex flex-row items-center gap-x-2">
            <Label for="allow-domain" class="font-normal text-gray-700 text-sm">
              Allow any user with a <b>@{$userDomain.data}</b> email address to
              join this project as a <b>Viewer</b>.
              <a
                target="_blank"
                href="https://docs.rilldata.com/reference/cli/user/whitelist"
              >
                Learn more
              </a>
            </Label>
            <div class="grow"></div>
            {#if $disallowDomainMutation.isLoading || $allowDomainMutation.isLoading}
              <LoadingCircleOutline />
            {:else}
              <Switch
                small
                checked={domainAllowed}
                id="allow-domain"
                class="mt-1"
                on:click={updateAllowedDomain}
              />
            {/if}
          </div>
        </div>
      {/if}

      <div class="mt-2 font-medium text-sm">
        {#if $allowedDomains.data?.domains?.length}
          <div>Domains added to allowlist by other admins</div>
          <div class="flex flex-col">
            {#each $allowedDomains.data.domains as { domain } (domain)}
              <div class="text-gray-500">@{domain}</div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </SettingsItemContainer>
{/if}
