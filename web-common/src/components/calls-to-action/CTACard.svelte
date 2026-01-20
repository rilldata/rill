<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  export let title: string;
  export let description: string | null = null;
  export let type: "primary" | "secondary" = "secondary";
  export let loading = false;
  export let disabled = false;
  export let ctaText: string;
  export let onClick: () => void;
</script>

<div
  class="flex flex-col gap-y-2 p-4 border rounded-lg text-base"
  class:bg-primary-50={type === "primary"}
>
  <div class="grow">
    <h3 class="text-base font-medium mb-2">{title}</h3>
    <div class="text-fg-secondary mb-2 text-sm">
      {#if $$slots.description}
        <slot name="description" />
      {:else if description}
        {description}
      {/if}
    </div>
  </div>
  <Button
    {type}
    {onClick}
    {loading}
    {disabled}
    forcedStyle="border-width: 1px !important; border-radius: calc(var(--radius) - 2px) !important;"
    large
  >
    {ctaText}
  </Button>
</div>
