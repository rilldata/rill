<script lang="ts">
  import { invalidate } from "$app/navigation";
  import {
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import UploadImagePopover from "@rilldata/web-admin/features/organizations/settings/UploadImagePopover.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let organization: string;
  export let organizationLogoUrl: string | undefined;
  export let organizationLogoDarkUrl: string | undefined;
  export let disabled = false;

  const logoUpdater = createAdminServiceUpdateOrganization({
    mutation: {
      mutationKey: ["updateOrganization", "logo", organization],
    },
  });
  $: ({
    error: logoError,
    isPending: isLogoLoading,
    mutateAsync: mutateLogoAsync,
  } = $logoUpdater);

  const logoDarkUpdater = createAdminServiceUpdateOrganization({
    mutation: {
      mutationKey: ["updateOrganization", "logoDark", organization],
    },
  });
  $: ({
    error: logoDarkError,
    isPending: isLogoDarkLoading,
    mutateAsync: mutateLogoDarkAsync,
  } = $logoDarkUpdater);

  async function onSaveLight(assetId: string) {
    await mutateLogoAsync({
      org: organization,
      data: {
        logoAssetId: assetId,
      },
    });
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    void invalidate("app:root");
  }

  async function onRemoveLight() {
    await mutateLogoAsync({
      org: organization,
      data: {
        logoAssetId: "",
      },
    });
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    void invalidate("app:root");
  }

  async function onSaveDark(assetId: string) {
    await mutateLogoDarkAsync({
      org: organization,
      data: {
        logoDarkAssetId: assetId,
      },
    });
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    void invalidate("app:root");
  }

  async function onRemoveDark() {
    await mutateLogoDarkAsync({
      org: organization,
      data: {
        logoDarkAssetId: "",
      },
    });
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    void invalidate("app:root");
  }

  $: hasAnyLogo = organizationLogoUrl || organizationLogoDarkUrl;
</script>

<SettingsContainer title="Logo" suppressFooter={!hasAnyLogo || disabled}>
  <div slot="body" class="flex flex-col gap-y-4">
    {#if disabled}
      <div class="relative">
        <div class="opacity-30 pointer-events-none">
          <div>
            Click to upload your logo and customize Rill for your organization.
          </div>
          <div class="flex flex-row gap-x-6 items-start mt-4">
            <div class="flex flex-col gap-y-2">
              <div class="text-sm font-medium">Light Logo</div>
              <div
                class="w-16 h-10 rounded border flex items-center justify-center"
              >
                <Rill width="64" height="40" mode="light" />
              </div>
            </div>
            <div class="flex flex-col gap-y-2">
              <div class="text-sm font-medium text-slate-500">Dark Logo</div>
              <div
                class="w-16 h-10 rounded border flex items-center justify-center"
              >
                <Rill width="64" height="40" mode="dark" />
              </div>
            </div>
          </div>
        </div>
        <div class="absolute inset-0 flex items-center justify-center">
          <a href="/{organization}/-/settings/billing">
            <Button type="primary">Upgrade to customize logo</Button>
          </a>
        </div>
      </div>
    {:else}
      <div>
        Click to upload your logo and customize Rill for your organization.
      </div>
      <div class="flex flex-row gap-x-6 items-start">
        <!-- Light Logo -->
        <div class="flex flex-col gap-y-2">
          <div class="text-sm font-medium">Light Logo</div>
          <UploadImagePopover
            imageUrl={organizationLogoUrl}
            accept="image/png, image/ico, image/x-ico, image/icon, image/x-icon"
            label="logo"
            {organization}
            loading={isLogoLoading}
            error={getRpcErrorMessage(logoError)}
            onSave={onSaveLight}
            onRemove={onRemoveLight}
          >
            <Rill width="64" height="40" mode="light" />
          </UploadImagePopover>
          {#if organizationLogoUrl}
            <Button
              type="secondary"
              onClick={onRemoveLight}
              loading={isLogoLoading}
              disabled={isLogoLoading}
              class="w-fit"
            >
              Remove
            </Button>
          {/if}
        </div>

        <!-- Dark Logo -->
        <div class="flex flex-col gap-y-2">
          <div class="text-sm font-medium">
            {#if organizationLogoDarkUrl}
              Dark Logo
            {:else}
              <span class="text-slate-500">Dark Logo</span>
            {/if}
          </div>
          <UploadImagePopover
            dark
            imageUrl={organizationLogoDarkUrl}
            accept="image/png, image/ico, image/x-ico, image/icon, image/x-icon"
            label="dark logo"
            {organization}
            loading={isLogoDarkLoading}
            error={getRpcErrorMessage(logoDarkError)}
            onSave={onSaveDark}
            onRemove={onRemoveDark}
          >
            <Rill width="64" height="40" mode="dark" />
          </UploadImagePopover>
          {#if organizationLogoDarkUrl}
            <Button
              type="secondary"
              onClick={onRemoveDark}
              loading={isLogoDarkLoading}
              disabled={isLogoDarkLoading}
              class="w-fit"
            >
              Remove
            </Button>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</SettingsContainer>
