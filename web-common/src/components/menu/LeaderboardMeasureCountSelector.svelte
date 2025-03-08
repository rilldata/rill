<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { fly } from "svelte/transition";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import IconButton from "../button/IconButton.svelte";

  export let measures: MetricsViewSpecMeasureV2[];
  export let count: number = 1;
  export let onMeasureCountChange: (count: number) => void;

  let isHovered = false;

  function handleIncrement() {
    if (count < measures.length) {
      onMeasureCountChange(Math.min(count + 1, measures.length));
    }
  }

  function handleDecrement() {
    if (count > 1) {
      onMeasureCountChange(Math.max(count - 1, 1));
    }
  }
</script>

<Button type="text">
  <div
    role="button"
    tabindex="0"
    class="flex items-center gap-x-1 px-1 font-normal"
    class:text-gray-700={!isHovered}
    class:text-inherit={isHovered}
    on:mouseenter={() => (isHovered = true)}
    on:mouseleave={() => (isHovered = false)}
  >
    {#if isHovered}
      <IconButton rounded on:click={handleDecrement} disabled={count <= 1}>
        <!-- TODO: Use a minus icon -->
        <svg
          width="14"
          height="14"
          viewBox="0 0 14 14"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M6.125 6.41602H2.1875C1.94588 6.41602 1.75 6.5466 1.75 6.70768V7.29102C1.75 7.4521 1.94588 7.58268 2.1875 7.58268H6.125L7.875 7.58268H11.8125C12.0541 7.58268 12.25 7.4521 12.25 7.29102V6.70768C12.25 6.5466 12.0541 6.41602 11.8125 6.41602H7.875H6.125Z"
            fill={count <= 1 ? "#94A3B8" : "#475569"}
          />
        </svg>
      </IconButton>
      <span class=" text-gray-700">
        <strong>{count} measure{count === 1 ? "" : "s"}</strong>
      </span>
      <IconButton
        rounded
        on:click={handleIncrement}
        disabled={count >= measures.length}
      >
        <!-- TODO: Use a plus icon -->
        <svg
          width="14"
          height="14"
          viewBox="0 0 14 14"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <g clip-path="url(#clip0_23394_553833)">
            <rect width="14" height="14" fill="#F8FAFC" />
            <path
              fill-rule="evenodd"
              clip-rule="evenodd"
              d="M6.41667 11.8125L6.41667 7.875V7.58333H6.125H2.1875C1.94588 7.58333 1.75 7.45275 1.75 7.29167V6.70833C1.75 6.54725 1.94588 6.41667 2.1875 6.41667H6.125H6.41667V6.125L6.41667 2.1875C6.41667 1.94588 6.54725 1.75 6.70833 1.75H7.29167C7.45275 1.75 7.58333 1.94588 7.58333 2.1875V6.125V6.41667H7.875H11.8125C12.0541 6.41667 12.25 6.54725 12.25 6.70833V7.29167C12.25 7.45275 12.0541 7.58333 11.8125 7.58333H7.875H7.58333V7.875V11.8125C7.58333 12.0541 7.45275 12.25 7.29167 12.25H6.70833C6.54725 12.25 6.41667 12.0541 6.41667 11.8125Z"
              fill="#334155"
            />
          </g>
          <defs>
            <clipPath id="clip0_23394_553833">
              <rect width="14" height="14" fill="white" />
            </clipPath>
          </defs>
        </svg>
      </IconButton>
    {:else}
      <span class="text-gray-700">
        Showing <strong>{count} measure{count === 1 ? "" : "s"}</strong>
      </span>
    {/if}
  </div>
</Button>
