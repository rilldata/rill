<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Rocket from "svelte-radix/Rocket.svelte";

  export let title: string;
  export let description: string;
  export let ctaLabel = "Deploy to Rill Cloud";
  export let ctaHref = "/deploy";
</script>

<div class="flex flex-col items-center gap-y-4 pt-16">
  <slot name="icon" />
  <div class="flex flex-col gap-y-2 items-center text-sm">
    <div class="text-fg-secondary font-semibold">
      {title}
    </div>
    <div class="text-fg-muted">
      {description}
    </div>
  </div>
  <div class="mt-4">
    <Button type="primary" href={ctaHref}>
      <Rocket size="16px" />
      {ctaLabel}
    </Button>
  </div>
</div>
