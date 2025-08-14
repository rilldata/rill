<script lang="ts">
  import { portal } from "@rilldata/web-common/lib/actions/portal";
  import type { NotificationMessage } from "@rilldata/web-common/lib/event-bus/events";
  import { onMount } from "svelte";
  import { scale } from "svelte/transition";
  import Button from "../button/Button.svelte";
  import Check from "../icons/Check.svelte";
  import Close from "../icons/Close.svelte";
  import LoadingSpinner from "../icons/LoadingSpinner.svelte";
  import WarningIcon from "../icons/WarningIcon.svelte";
  import { NOTIFICATION_TIMEOUT } from "./constants";

  export let location: "top" | "bottom" | "middle" = "bottom";
  export let justify: "left" | "right" | "center" = "center";
  export let notification: NotificationMessage;
  export let onClose: () => void;

  $: ({ message, link, type, detail, options } = notification);

  onMount(() => {
    if (!options?.persisted && !link && type !== "loading") {
      const timeout = options?.timeout ?? NOTIFICATION_TIMEOUT;
      setTimeout(onClose, timeout);
    }
  });
</script>

<aside
  use:portal
  transition:scale={{ duration: 200, start: 0.98, opacity: 0 }}
  class="{location} {justify}"
  aria-label="Notification"
>
  <div class="main-section">
    <div class="message-container" class:font-medium={detail}>
      {#if type === "success"}
        <Check size="18px" className="text-gray-800" />
      {:else if type === "loading"}
        <LoadingSpinner size="18px" />
      {:else if type == "error"}
        <WarningIcon />
      {/if}

      {message}
    </div>

    {#if link}
      <div class="link-container">
        <a href={link.href} on:click={onClose} class="text-secondary-400">
          {link.text}
        </a>
      </div>
    {/if}

    {#if options?.persisted && type !== "loading"}
      <div class="px-2 py-2 border-l">
        <Button onClick={onClose} square>
          <Close size="18px" color="#fff" />
        </Button>
      </div>
    {/if}
  </div>

  {#if detail}
    <hr />
    <div class="detail">
      {detail}
    </div>
  {/if}
</aside>

<style lang="postcss">
  * {
    @apply border-gray-600;
  }

  aside {
    @apply absolute w-fit z-50 flex flex-col text-sm;
    @apply bg-gray-800 text-gray-200 p-0 rounded-md shadow-lg;
  }

  .main-section {
    @apply flex;
  }

  .detail {
    @apply px-4 pt-2 pb-3;
  }

  .link-container {
    @apply px-4 py-2 border-l items-center flex;
  }

  .message-container {
    @apply flex items-center px-4 py-2 gap-x-2;
  }

  a {
    @apply text-primary-300;
  }

  a:hover {
    @apply text-primary-200;
  }

  .top {
    @apply top-[30px];
  }

  .bottom {
    @apply bottom-[30px];
  }

  .middle {
    @apply top-1/2;
    @apply -translate-y-1/2;
  }

  .left {
    @apply left-4;
  }

  .right {
    @apply right-4;
  }

  .center {
    @apply left-1/2;
    @apply -translate-x-1/2;
  }
</style>
