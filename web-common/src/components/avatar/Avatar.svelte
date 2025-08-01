<script lang="ts">
  import { Avatar } from "bits-ui";
  import { cn } from "@rilldata/web-common/lib/shadcn";

  export let loadingStatus: Avatar.Props["loadingStatus"] = undefined;
  export let src: string | null = null;
  export let alt: string | null = null;
  export let avatarSize: string = "h-12 w-12";
  export let fontSize: string = "text-xs";
  export let bgColor: string = "bg-blue-500";

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
</script>

<Avatar.Root
  bind:loadingStatus
  class={cn(
    avatarSize,
    "rounded-full",
    loadingStatus === "loaded" ? "border-foreground" : "border-transparent",
    "text-[17px]",
    "font-medium",
    "uppercase",
  )}
  aria-label="User avatar"
>
  <div
    class={cn(
      avatarSize,
      `flex items-center justify-center overflow-hidden rounded-full border`,
      {
        "border-dashed bg-transparent border-slate-400": !src && !alt,
        [`border-transparent ${bgColor}`]:
          (!src && alt) || (loadingStatus === "error" && alt),
      },
    )}
  >
    {#if src}
      <Avatar.Image {src} {alt} />
      {#if alt}
        <!-- Show a fallback if the image fails to load -->
        <Avatar.Fallback class={cn(fontSize, "text-white")}>
          {getInitials(alt ?? "")}
        </Avatar.Fallback>
      {/if}
    {:else if alt}
      <Avatar.Fallback class={cn(fontSize, "text-white")}>
        {getInitials(alt)}
      </Avatar.Fallback>
    {:else}
      <Avatar.Fallback class={cn(fontSize, "text-slate-400")}>
        <svg
          class="mt-[6px]"
          width="24"
          height="22"
          viewBox="0 0 24 22"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M12 10C14.7614 10 17 7.76142 17 5C17 2.23858 14.7614 0 12 0C9.23858 0 7 2.23858 7 5C7 7.76142 9.23858 10 12 10Z"
            fill="#94A3B8"
          />
          <path
            d="M12 33.913C18.6274 33.913 24 29.0076 24 22.9565C24 16.9054 18.6274 12 12 12C5.37258 12 0 16.9054 0 22.9565C0 29.0076 5.37258 33.913 12 33.913Z"
            fill="#94A3B8"
          />
        </svg>
      </Avatar.Fallback>
    {/if}
  </div>
</Avatar.Root>
