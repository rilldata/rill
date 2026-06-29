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
    organizationThumbnailUrl,
  }: {
    organization: string;
    organizationThumbnailUrl: string | undefined;
  } = $props();

  const orgUpdater = createAdminServiceUpdateOrganization();
  let { error, isPending: isLoading, mutateAsync } = $derived($orgUpdater);

  async function onSave(assetId: string) {
    await mutateAsync({
      org: organization,
      data: {
        thumbnailAssetId: assetId,
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
        thumbnailAssetId: "",
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
  title={m.settings_thumbnail_title()}
  action={organizationThumbnailUrl ? removeAction : undefined}
>
  <div class="flex flex-col gap-y-2">
    <div>
      {m.settings_thumbnail_description()}
    </div>
    <UploadImagePopover
      imageUrl={organizationThumbnailUrl}
      accept="image/png, image/jpeg, image/gif"
      label="thumbnail"
      {organization}
      loading={isLoading}
      error={getRpcErrorMessage(error)}
      {onSave}
      {onRemove}
    >
      <img
        src="https://cdn.rilldata.com/images/rill-admin.png"
        alt="thumbnail"
        class="h-10"
      />
    </UploadImagePopover>
  </div>
</SettingsContainer>
