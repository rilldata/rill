<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import IconButton from "../button/IconButton.svelte";
  import MinusIcon from "../icons/MinusIcon.svelte";
  import PlusIcon from "../icons/PlusIcon.svelte";

  export let measures: MetricsViewSpecMeasureV2[];
  export let count: number = 1;
  export let onMeasureCountChange: (count: number) => void;

  let isHovered = false;

  $: maxCount = measures.length;

  function handleIncrement() {
    if (count < maxCount) {
      onMeasureCountChange(Math.min(count + 1, maxCount));
    }
  }

  function handleDecrement() {
    if (count > 1) {
      onMeasureCountChange(Math.max(count - 1, 1));
    }
  }

  $: visibleMeasuresCount = measures.length;

  // Update count to match visible measures when filtered count is greater
  $: if (visibleMeasuresCount < count) {
    onMeasureCountChange(visibleMeasuresCount);
  }

  $: copyText = isHovered
    ? `${Math.min(count, visibleMeasuresCount)} measure${Math.min(count, visibleMeasuresCount) === 1 ? "" : "s"}`
    : `Showing ${Math.min(count, visibleMeasuresCount)} measure${Math.min(count, visibleMeasuresCount) === 1 ? "" : "s"}`;
</script>

<Button type="text" forcedStyle="width: 133px;">
  <div
    role="button"
    tabindex="0"
    class="flex items-center gap-x-2 font-normal h-6"
    on:mouseenter={() => (isHovered = true)}
    on:mouseleave={() => (isHovered = false)}
  >
    <div class="flex items-center gap-x-2 min-w-[120px]">
      {#if isHovered}
        <IconButton rounded on:click={handleDecrement} disabled={count <= 1}>
          <MinusIcon size="14" color={count <= 1 ? "#94A3B8" : "#475569"} />
        </IconButton>
        <span class="text-gray-700">
          <strong>{copyText}</strong>
        </span>
        <IconButton
          rounded
          on:click={handleIncrement}
          disabled={count >= measures.length}
        >
          <PlusIcon
            size="14"
            color={count >= measures.length ? "#94A3B8" : "#475569"}
          />
        </IconButton>
      {:else}
        <span class="text-gray-700">
          <strong>{copyText}</strong>
        </span>
      {/if}
    </div>
  </div>
</Button>
