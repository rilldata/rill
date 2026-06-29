<script lang="ts">
  import IconSwitcher from "@rilldata/web-common/components/forms/IconSwitcher.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { ArrowDown, ArrowRight } from "lucide-svelte";

  export let key: string;
  export let label: string;
  export let value: string | undefined;
  export let onChange: (updatedSparkline: string) => void;

  $: horizontalOptions = [
    {
      id: "bottom",
      Icon: ArrowDown,
      tooltip: m.canvas_sparkline_below(),
    },
    {
      id: "right",
      Icon: ArrowRight,
      tooltip: m.canvas_sparkline_right(),
    },
  ];

  $: showSparkline = value !== "none";
  $: isSparkRight = value === "right";
</script>

<div class="flex flex-col gap-y-2">
  <div class="flex justify-between py-1 items-center">
    <InputLabel small {label} id={key} faint={!showSparkline} />
    <Switch
      checked={showSparkline}
      onclick={() => {
        let newSparklinePosition = "bottom";
        if (showSparkline) newSparklinePosition = "none";
        onChange(newSparklinePosition);
      }}
      small
    />
  </div>

  {#if showSparkline}
    <IconSwitcher
      small
      expand
      fields={horizontalOptions}
      selected={isSparkRight ? "right" : "bottom"}
      onClick={(option) => onChange(option)}
    />
  {/if}
</div>
