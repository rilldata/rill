<script lang="ts">
  import { Avatar } from "bits-ui";

  export let loadingStatus: Avatar.Props["loadingStatus"] = undefined;
  export let src: string | null = null;
  export let alt: string;
  export let size: string = "h-12 w-12";
</script>

<Avatar.Root
  bind:loadingStatus
  class="{size} rounded-full border {loadingStatus === 'loaded'
    ? 'border-foreground'
    : 'border-transparent'} bg-muted text-[17px] font-medium uppercase text-muted-foreground"
>
  <div
    class="flex h-full w-full items-center justify-center overflow-hidden rounded-full border-2 border-transparent"
  >
    {#if src}
      <Avatar.Image {src} {alt} />
      <Avatar.Fallback class="border border-muted text-xs">
        {alt.charAt(0)}
      </Avatar.Fallback>
    {:else}
      <Avatar.Fallback class="border border-muted text-xs">
        {alt.charAt(0)}
      </Avatar.Fallback>
    {/if}
  </div>
</Avatar.Root>
