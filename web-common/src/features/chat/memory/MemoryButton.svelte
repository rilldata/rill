<!--
  Opens the memory manager. Renders as an icon button in chat headers, as a full-width
  secondary button in the fullpage conversation sidebar footer, and as a square secondary
  button when that sidebar is collapsed, matching the "Connect your own client" button
  next to it. Hidden when memory is off for the project.
-->
<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Brain } from "lucide-svelte";
  import MemoryDialog from "./MemoryDialog.svelte";

  export let variant: "icon" | "button" | "square" = "icon";
  export let conversationBasePath: string | undefined = undefined;

  const { chatMemory } = featureFlags;

  let open = false;
</script>

{#if $chatMemory}
  {#if variant === "icon"}
    <IconButton
      ariaLabel={m.chat_memory_open()}
      bgGray
      active={open}
      disableTooltip={open}
      onclick={() => (open = true)}
    >
      <Brain size="16px" class="text-fg-muted" />
      <svelte:fragment slot="tooltip-content">
        {m.chat_memory_open()}
      </svelte:fragment>
    </IconButton>
  {:else if variant === "button"}
    <Button type="secondary" class="w-full" onClick={() => (open = true)}>
      <Brain size="14px" />
      {m.chat_memory_open()}
    </Button>
  {:else}
    <span title={m.chat_memory_open()}>
      <Button
        type="secondary"
        square
        label={m.chat_memory_open()}
        onClick={() => (open = true)}
      >
        <Brain size="14px" />
      </Button>
    </span>
  {/if}
  <MemoryDialog bind:open {conversationBasePath} />
{/if}
