<script lang="ts">
  import { Tag } from "@rilldata/web-common/components/tag";
  import type { Color } from "@rilldata/web-common/components/tag/Tag.svelte";
  import {
    V1AlertExecution,
    V1AssertionStatus,
    type V1AssertionResult,
  } from "@rilldata/web-common/runtime-client";

  export let currentExecution: V1AlertExecution;
  export let result: V1AssertionResult;

  type AssertionResultDisplay = {
    text: string;
    color: Color;
  };
  const assertionResultDisplays: Record<
    V1AssertionStatus,
    AssertionResultDisplay
  > = {
    [V1AssertionStatus.ASSERTION_STATUS_PASS]: {
      text: "Not triggered",
      color: "gray",
    },
    [V1AssertionStatus.ASSERTION_STATUS_FAIL]: {
      text: "Triggered",
      color: "blue",
    },
    [V1AssertionStatus.ASSERTION_STATUS_ERROR]: {
      text: "Failed",
      color: "red",
    },
    // This should never happen
    [V1AssertionStatus.ASSERTION_STATUS_UNSPECIFIED]: {
      text: "Status unknown",
      color: "amber",
    },
  };

  $: assertionResultDisplay = assertionResultDisplays[result.status];
</script>

{#if currentExecution}
  <Tag color="green">Running</Tag>
{:else}
  <Tag color={assertionResultDisplay.color}>
    {assertionResultDisplay.text}
  </Tag>
{/if}
