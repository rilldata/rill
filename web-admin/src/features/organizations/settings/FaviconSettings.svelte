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
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let organization: string;
  export let organizationFaviconUrl: string | undefined;
  export let disabled = false;

  const orgUpdater = createAdminServiceUpdateOrganization();
  $: ({ error, isPending: isLoading, mutateAsync } = $orgUpdater);

  async function onSave(assetId: string) {
    await mutateAsync({
      org: organization,
      data: {
        faviconAssetId: assetId,
      },
    });
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    void invalidate("app:root");
  }

  async function onRemove() {
    await mutateAsync({
      org: organization,
      data: {
        faviconAssetId: "",
      },
    });
    void queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    void invalidate("app:root");
  }
</script>

<SettingsContainer
  title="Favicon"
  suppressFooter={!organizationFaviconUrl || disabled}
>
  <div slot="body" class="flex flex-col gap-y-2">
    {#if disabled}
      <div class="relative">
        <div class="opacity-30 pointer-events-none">
          <div>
            Click to upload your favicon and customize Rill for your
            organization. Upload a square icon to get the best results.
          </div>
          <div class="mt-2">
            <img src="/favicon.png" alt="favicon" class="h-10" />
          </div>
        </div>
        <div class="absolute inset-0 flex items-center justify-center">
          <a href="/{organization}/-/settings/billing">
            <Button type="primary">Upgrade to customize favicon</Button>
          </a>
        </div>
      </div>
    {:else}
      <div>
        Click to upload your favicon and customize Rill for your organization.
        Upload a square icon to get the best results.
      </div>
      <UploadImagePopover
        imageUrl={organizationFaviconUrl}
        accept="image/png, image/ico, image/x-ico, image/icon, image/x-icon"
        label="favicon"
        {organization}
        loading={isLoading}
        error={getRpcErrorMessage(error)}
        {onSave}
        {onRemove}
      >
        <img src="/favicon.png" alt="favicon" class="h-10" />
      </UploadImagePopover>
    {/if}
  </div>
  <svelte:fragment slot="action">
    {#if organizationFaviconUrl && !disabled}
      <Button
        type="secondary"
        onClick={onRemove}
        loading={isLoading}
        disabled={isLoading}
      >
        Remove
      </Button>
    {/if}
  </svelte:fragment>
</SettingsContainer>
