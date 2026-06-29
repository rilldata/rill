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
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let {
    organization,
    organizationFaviconUrl,
  }: {
    organization: string;
    organizationFaviconUrl: string | undefined;
  } = $props();

  const orgUpdater = createAdminServiceUpdateOrganization();
  let { error, isPending: isLoading, mutateAsync } = $derived($orgUpdater);

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

{#snippet removeAction()}
  <Button
    type="secondary"
    onClick={onRemove}
    loading={isLoading}
    disabled={isLoading}
  >
    {m.settings_remove_button()}
  </Button>
{/snippet}

<SettingsContainer
  title={m.settings_favicon_title()}
  action={organizationFaviconUrl ? removeAction : undefined}
>
  <div class="flex flex-col gap-y-2">
    <div>
      {m.settings_favicon_description()}
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
  </div>
</SettingsContainer>
