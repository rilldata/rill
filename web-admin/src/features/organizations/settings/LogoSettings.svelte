<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import {
    createAdminServiceCreateAsset,
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { CANONICAL_ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Rill from "@rilldata/web-common/components/icons/Rill.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { builderActions, getAttrs } from "bits-ui";
  import FileInput from "@rilldata/web-common/components/forms/FileInput.svelte";

  export let organization: string;
  export let organizationLogoUrl: string | undefined;

  $: logoUrl = organizationLogoUrl;

  const assetCreator = createAdminServiceCreateAsset();
  const orgUpdater = createAdminServiceUpdateOrganization();

  let open = false;
  let assetId = "";

  async function uploadFile(file: File) {
    const ext = extractFileExtension(file.name);
    const assetResp = await $assetCreator.mutateAsync({
      organizationName: organization,
      data: {
        type: "image",
        name: "logo",
        extension: ext,
        public: true,
        estimatedSizeBytes: file.size.toString(),
      },
    });

    await fetch(assetResp.signedUrl, {
      method: "PUT",
      body: file,
      headers: assetResp.signingHeaders,
    });
    assetId = assetResp.assetId;
    return `${CANONICAL_ADMIN_URL}/v1/assets/${assetId}/download`;
  }

  function onCancel() {
    open = false;
    logoUrl = organizationLogoUrl;
  }

  async function removeLogo() {
    onCancel();
    await $orgUpdater.mutateAsync({
      name: organization,
      data: {
        logoAssetId: "",
      },
    });
    void queryClient.invalidateQueries(
      getAdminServiceGetOrganizationQueryKey(organization),
    );
    void invalidateAll();
  }

  async function onSave() {
    onCancel();
    await $orgUpdater.mutateAsync({
      name: organization,
      data: {
        logoAssetId: assetId,
      },
    });
    void queryClient.invalidateQueries(
      getAdminServiceGetOrganizationQueryKey(organization),
    );
    void invalidateAll();
  }
</script>

<SettingsContainer title="Logo" suppressFooter={!organizationLogoUrl}>
  <div slot="body" class="flex flex-col gap-y-2">
    <div>
      Click to upload your logo and customize Rill for your organization.
    </div>
    <Popover
      bind:open
      onOpenChange={(o) => {
        if (!o) onCancel();
      }}
    >
      <PopoverTrigger asChild let:builder>
        <button
          class="flex items-center relative group h-[72px] border border-gray-300 hover:bg-slate-100 w-fit"
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
          class:w-24={!organizationLogoUrl}
          class:w-20={!!organizationLogoUrl}
        >
          <div class="m-auto px-4 w-fit h-10">
            {#if organizationLogoUrl}
              <img src={organizationLogoUrl} alt="logo" class="h-10" />
            {:else}
              <Rill width="64" height="40" />
            {/if}
          </div>
          {#if !open}
            <div
              class="absolute -bottom-2 -right-2 rounded-2xl bg-slate-200 group-hover:bg-slate-500 w-6 h-6 px-1.5 py-[5px]"
            >
              <EditIcon
                size="16px"
                className="text-slate-600 group-hover:text-slate-50"
              />
            </div>
          {/if}
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        side="bottom"
        class="flex flex-col gap-y-2 w-[400px] p-4"
      >
        <div class="text-base font-medium">Upload org logo</div>
        <FileInput bind:value={logoUrl} accept="image/*" {uploadFile} />
        <div class="flex flex-row justify-end gap-x-2">
          <Button type="secondary" on:click={onCancel}>Cancel</Button>
          {#if organizationLogoUrl}
            <Button type="secondary" on:click={removeLogo}>Remove</Button>
          {/if}
          <Button type="primary" on:click={onSave}>Save</Button>
        </div>
      </PopoverContent>
    </Popover>
  </div>
  <svelte:fragment slot="action">
    {#if organizationLogoUrl}
      <Button type="secondary" on:click={removeLogo}>Remove</Button>
    {/if}
  </svelte:fragment>
</SettingsContainer>
