<!--
  A small pill after a measure name that shows whether the measure is computed
  with its approximate default (≈) or its exact expression (=), and toggles
  between the two on click. A canvas chart has one chip for all its measures,
  next to its title. Every host is a link, a sort button or a header
  that reacts to mouse events, so the chip stops its events from propagating.
  Hosts with a tooltip of their own hide it while the chip is hovered, through
  `onHoverChange`, so that only the chip's tooltip shows.
-->
<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { PrecisionChipProps } from "./exact-measures";

  // `state` is renamed locally: a binding called `state` conflicts with the `$state` rune.
  let {
    state: precision,
    measureLabel,
    loading = false,
    error = undefined,
    onToggle = undefined,
    onHoverChange = undefined,
  }: PrecisionChipProps = $props();

  const canToggle = $derived(!!onToggle);

  const tooltip = $derived.by(() => {
    if (loading) return m.dashboard_precision_loading();
    if (error) return m.dashboard_precision_error({ message: error });
    switch (precision) {
      case "exact":
        return m.dashboard_precision_exact();
      case "inherited":
        return m.dashboard_precision_inherited();
      case "partial":
        return m.dashboard_precision_partial();
      default:
        return m.dashboard_precision_approx();
    }
  });

  let hovered = $state(false);

  function setHovered(value: boolean) {
    if (hovered === value) return;
    hovered = value;
    onHoverChange?.(value);
  }

  // A chip removed while hovered must not leave its host's tooltip hidden.
  $effect(() => () => {
    if (hovered) onHoverChange?.(false);
  });

  function toggle(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    if (canToggle) onToggle?.();
  }
</script>

<Tooltip location="top" distance={6} activeDelay={150}>
  <button
    type="button"
    class="precision-chip"
    class:exact={precision === "exact" && !error}
    class:inherited={(precision === "inherited" || precision === "partial") &&
      !error}
    class:error={!!error}
    class:interactive={canToggle}
    aria-pressed={precision === "partial" ? "mixed" : precision === "exact"}
    aria-label={m.dashboard_precision_toggle_aria({ measure: measureLabel })}
    aria-busy={loading}
    onclick={toggle}
    onmousedown={(e) => e.stopPropagation()}
    onmouseenter={() => setHovered(true)}
    onmouseleave={() => setHovered(false)}
    onfocus={() => setHovered(true)}
    onblur={() => setHovered(false)}
  >
    {#if loading}
      <Spinner size="9px" status={EntityStatus.Running} />
    {:else if error}
      !
    {:else if precision === "approx"}
      ≈
    {:else}
      =
    {/if}
  </button>
  <TooltipContent slot="tooltip-content" maxWidth="260px">
    {tooltip}
  </TooltipContent>
</Tooltip>

<style lang="postcss">
  .precision-chip {
    @apply inline-flex items-center justify-center shrink-0;
    @apply h-[15px] min-w-[15px] px-1 rounded-full;
    @apply border border-border text-fg-muted bg-surface-background;
    @apply text-[10px] font-bold leading-none not-italic;
    @apply cursor-default;
  }
  .precision-chip.interactive {
    @apply cursor-pointer;
  }
  .precision-chip.interactive:hover {
    @apply border-fg-muted text-fg;
  }
  .precision-chip.exact {
    @apply bg-primary-100 border-primary-400 text-primary-700;
  }
  .precision-chip.exact.interactive:hover {
    @apply border-primary-600 text-primary-800;
  }
  .precision-chip.inherited {
    @apply bg-primary-50 border-primary-300 text-primary-700 border-dashed;
  }
  .precision-chip.inherited.interactive:hover {
    @apply border-primary-500 text-primary-800;
  }
  .precision-chip.error {
    @apply bg-red-50 border-red-300 text-red-700;
  }
</style>
