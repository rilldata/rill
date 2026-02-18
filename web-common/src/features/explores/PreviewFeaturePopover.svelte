<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import { Button } from "../../components/button";
  import Rocket from "svelte-radix/Rocket.svelte";

  export let title: string;
  export let description: string;
  export let buttonLabel: string;
  export let compact = true;
  export let gray = true;
</script>

<Popover.Root>
  <Popover.Trigger asChild let:builder>
    <Button
      type="secondary"
      {compact}
      {gray}
      builders={[builder]}
      label={buttonLabel}
    >
      <slot name="icon" />
      <slot name="trigger-label" />
    </Button>
  </Popover.Trigger>
  <Popover.Content align="end" class="w-64">
    <div class="flex flex-col gap-y-3">
      <div>
        <p class="text-sm font-semibold" style="color: var(--fg-primary)">
          {title}
        </p>
        <p class="text-xs mt-1" style="color: var(--fg-muted)">
          {description}
        </p>
      </div>
      <Button type="primary" href="/deploy" compact>
        <Rocket size="14px" />
        Deploy to unlock
      </Button>
    </div>
  </Popover.Content>
</Popover.Root>
