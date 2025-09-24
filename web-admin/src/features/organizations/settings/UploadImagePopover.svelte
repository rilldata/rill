<script lang="ts">
  import { createAdminServiceCreateAsset } from "@rilldata/web-admin/client";
  import { CANONICAL_ADMIN_URL } from "@rilldata/web-admin/client/http-client";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import FileInput from "@rilldata/web-common/components/forms/FileInput.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover/index.js";
  import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
  import { builderActions, getAttrs } from "bits-ui";

  export let imageUrl: string;
  export let accept: string;
  export let label: string;
  export let organization: string;
  export let loading: boolean;
  export let error: string;
  export let onSave: (assetId: string) => Promise<void>;
  export let onRemove: () => Promise<void>;

  // `imageUrl` is the saved image while `url` is the temporarily uploaded image.
  // Since save only happens when `Save` is clicked these could be different.
  $: url = imageUrl;

  let open = false;
  let assetId = "";

  const assetCreator = createAdminServiceCreateAsset();

  async function uploadFile(file: File) {
    const ext = extractFileExtension(file.name);
    const assetResp = await $assetCreator.mutateAsync({
      org: organization,
      data: {
        type: "image",
        name: label,
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
    url = imageUrl;
  }

  async function handleRemove() {
    await onRemove();
    onCancel();
  }

  async function handleSave() {
    await onSave(assetId);
    onCancel();
  }
</script>

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
      class:w-24={!imageUrl}
      class:w-20={!!imageUrl}
    >
      <div class="m-auto px-4 w-fit h-10">
        {#if imageUrl}
          <img src={imageUrl} alt={label} class="h-10" />
        {:else}
          <slot />
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
    <div class="text-base font-medium">Upload org {label}</div>
    <FileInput bind:value={url} {accept} {uploadFile} />
    {#if error}
      <div class="text-red-600 text-xs">
        {error}
      </div>
    {/if}
    <div class="flex flex-row justify-end gap-x-2">
      <Button type="secondary" onClick={onCancel}>Cancel</Button>
      {#if imageUrl}
        <Button
          type="secondary"
          onClick={handleRemove}
          {loading}
          disabled={loading}
        >
          Remove
        </Button>
      {/if}
      <Button
        type="primary"
        onClick={handleSave}
        {loading}
        disabled={loading || !assetId}
      >
        Save
      </Button>
    </div>
  </PopoverContent>
</Popover>
