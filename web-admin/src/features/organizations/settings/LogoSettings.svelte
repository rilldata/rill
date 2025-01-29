<script lang="ts">
  import {
    createAdminServiceCreateAsset,
    createAdminServiceGetOrganization,
  } from "@rilldata/web-admin/client";
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
  import { builderActions, getAttrs } from "bits-ui";
  import FileInput from "@rilldata/web-common/components/forms/FileInput.svelte";

  export let organization: string;

  $: orgResp = createAdminServiceGetOrganization(organization);
  $: logoUrl = $orgResp.data?.organization?.logoUrl;

  const assetCreator = createAdminServiceCreateAsset();

  let open = false;

  async function uploadFile(file: File) {
    const ext = extractFileExtension(file.name);
    const assetResp = await $assetCreator.mutateAsync({
      organizationName: organization,
      data: {
        type: "image",
        name: "logo",
        extension: ext,
        cacheable: true,
        estimatedSizeBytes: file.size,
      },
    });

    const formData = new FormData();
    formData.append("file", file);
    const resp = await fetch(assetResp.signedUrl, {
      method: "PUT",
      body: formData,
      headers: assetResp.signingHeaders,
    });
    console.log(resp.statusText);
  }
</script>

<SettingsContainer title="Logo" suppressFooter={!logoUrl}>
  <div slot="body" class="flex flex-col gap-y-2">
    <div>
      Click to upload your logo and customize Rill for your organization.
    </div>
    <Popover bind:open>
      <PopoverTrigger asChild let:builder>
        <button
          class="flex items-center relative group h-[72px] border border-gray-300 hover:bg-slate-100"
          {...getAttrs([builder])}
          use:builderActions={{ builders: [builder] }}
          class:w-24={!logoUrl}
          class:w-20={!!logoUrl}
        >
          <div class="m-auto w-fit h-10">
            {#if logoUrl}
              <img src={logoUrl} alt="logo" class="h-10" />
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
      <PopoverContent align="start" class="w-[400px] p-4">
        <div>Upload org logo</div>
        <FileInput bind:value={logoUrl} accept="image/*" {uploadFile} />
        <div>
          <Button type="secondary" on:click={() => (open = false)}>
            Cancel
          </Button>
          {#if logoUrl}
            <Button type="secondary">Remove</Button>
          {/if}
          <Button type="primary">Save</Button>
        </div>
      </PopoverContent>
    </Popover>
  </div>
  <svelte:fragment slot="action">
    {#if logoUrl}
      <Button type="secondary">Remove</Button>
    {/if}
  </svelte:fragment>
</SettingsContainer>
