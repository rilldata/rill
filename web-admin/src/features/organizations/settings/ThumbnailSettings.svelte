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
  export let organizationThumbnailUrl: string | undefined;

  const orgUpdater = createAdminServiceUpdateOrganization();
  $: ({ error, isPending: isLoading, mutateAsync } = $orgUpdater);

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

<SettingsContainer title="Thumbnail" suppressFooter={!organizationThumbnailUrl}>
  <div slot="body" class="flex flex-col gap-y-2">
    <div>
      Click to upload your thumbnail. The thumbnail will be used when sharing
      links to Rill in applications like Slack.
    </div>
    <UploadImagePopover
      imageUrl={organizationThumbnailUrl}
      accept="image/png, image/jpeg, image/gif, image/svg+xml"
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
  <svelte:fragment slot="action">
    {#if organizationThumbnailUrl}
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
