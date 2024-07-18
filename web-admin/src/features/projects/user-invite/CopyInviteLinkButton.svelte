<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { isClipboardApiSupported } from "@rilldata/web-common/lib/actions/copy-to-clipboard";

  export let copyLink: string;

  let copied = false;
  function onCopy() {
    navigator.clipboard.writeText(copyLink).catch(console.error);
    copied = true;
  }
</script>

{#if isClipboardApiSupported()}
  {#if copied}
    <div class="flex flex-row gap-x-1 items-center">
      <Check size="12px" />
      <span class="font-medium text-xs">Link copied</span>
    </div>
  {:else}
    <Button
      type="link"
      class="flex flex-row items-center"
      forcedStyle="min-height: 24px !important; height: 24px !important; padding-right: 0px !important;"
      on:click={onCopy}
      compact
    >
      <Link size="12px" />
      <span class="font-medium text-xs">Copy link</span>
    </Button>
  {/if}
{/if}
