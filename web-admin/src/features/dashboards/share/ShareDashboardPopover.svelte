<script lang="ts">
  import CreatePublicURLForm from "@rilldata/web-admin/features/public-urls/CreatePublicURLForm.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";

  export let createMagicAuthTokens: boolean;

  const { hidePublicUrl } = featureFlags;
  let isOpen = false;
  let copied = false;

  function onCopy() {
    navigator.clipboard.writeText(window.location.href).catch(console.error);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }
</script>

<Popover
  bind:open={isOpen}
  onOutsideClick={() => {
    isOpen = false;
  }}
>
  <PopoverTrigger asChild let:builder>
    <Button type="secondary" builders={[builder]} selected={isOpen}
      >Share</Button
    >
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">Copy URL</TabsTrigger>
        {#if createMagicAuthTokens && !$hidePublicUrl}
          <TabsTrigger value="tab2">Create public URL</TabsTrigger>
        {/if}
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-4">
          <h3 class="text-xs text-gray-800 font-normal">
            Share your current view with another project member.
          </h3>
          <Button
            type="secondary"
            onClick={() => {
              onCopy();
            }}
          >
            {#if copied}
              <Check size="16px" />
              Copied URL
            {:else}
              <Link size="16px" className="text-primary-500" />
              Copy URL for this view
            {/if}
          </Button>
        </div>
      </TabsContent>
      <TabsContent value="tab2" class="mt-0 p-4">
        {#if createMagicAuthTokens && !$hidePublicUrl}
          <CreatePublicURLForm />
        {/if}
      </TabsContent>
    </Tabs>
  </PopoverContent>
</Popover>

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
