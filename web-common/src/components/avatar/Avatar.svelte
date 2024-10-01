<script lang="ts">
  import { Avatar } from "bits-ui";
  import { UserRoundIcon } from "lucide-svelte";

  export let loadingStatus: Avatar.Props["loadingStatus"] = undefined;
  export let src: string | null = null;
  export let alt: string | null = null;
  export let size: string = "h-12 w-12";

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
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
    {#if !src}
      {#if alt}
        <Avatar.Image {src} {alt} />
        <Avatar.Fallback class="border border-muted text-xs">
          {getInitials(alt)}
        </Avatar.Fallback>
      {:else}
        <Avatar.Fallback class="border-dashed border-muted text-xs">
          <UserRoundIcon size="20px" class="mt-1" />
        </Avatar.Fallback>
      {/if}
    {/if}
  </div>
</Avatar.Root>
