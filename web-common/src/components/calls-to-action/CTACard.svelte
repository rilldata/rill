<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import type { ButtonType } from "../button/Button.svelte";

  export let title: string;
  export let description: string | null = null;
  export let type: ButtonType = "secondary";
  export let loading = false;
  export let disabled = false;
  export let ctaText: string;
  export let onClick: () => void;
</script>

<div class="flex flex-col gap-y-2 p-4 border rounded-lg text-base">
  <div class="grow">
    <h3 class="text-base font-medium mb-2">{title}</h3>
    <div class="text-fg-muted mb-2 text-sm">
      {#if $$slots.description}
        <slot name="description" />
      {:else if description}
        {description}
      {/if}
    </div>
  </div>
  <Button {type} {onClick} {loading} {disabled} large>
    {ctaText}
  </Button>
</div>
