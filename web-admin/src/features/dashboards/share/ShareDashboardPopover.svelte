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
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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

<Popover bind:open={isOpen}>
  <PopoverTrigger>
    {#snippet child({ props })}
      <Tooltip distance={8} suppress={isOpen}>
        <Button {...props} type="secondary" selected={isOpen}
          >{m.avatar_share()}</Button
        >
        <TooltipContent slot="tooltip-content"
          >{m.avatar_share_dashboard()}</TooltipContent
        >
      </Tooltip>
    {/snippet}
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">{m.avatar_copy_url()}</TabsTrigger>
        {#if createMagicAuthTokens && !$hidePublicUrl}
          <TabsTrigger value="tab2">{m.avatar_create_public_url()}</TabsTrigger>
        {/if}
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-4">
          <h3 class="text-xs text-fg-primary font-normal">
            {m.avatar_share_description()}
          </h3>
          <Button
            type="secondary"
            onClick={() => {
              onCopy();
            }}
          >
            {#if copied}
              <Check size="16px" />
              {m.avatar_copied_url()}
            {:else}
              <Link size="16px" className="text-primary-500" />
              {m.avatar_copy_url_for_view()}
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
