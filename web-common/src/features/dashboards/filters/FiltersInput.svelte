<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import {
    convertExpressionToFilterParam,
    convertFilterParamToExpression,
    stripParserError,
  } from "@rilldata/web-common/features/dashboards/url-state/filters/converters";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { InfoIcon } from "lucide-svelte";

  export let filter: V1Expression;
  export let onChange: (newFilter: V1Expression) => void;

  let filterText = convertExpressionToFilterParam(filter);
  let error = "";

  function handleFilterChange() {
    try {
      const newFilter = convertFilterParamToExpression(filterText);
      if (newFilter) {
        onChange(newFilter);
        error = "";
      }
    } catch (e) {
      error = stripParserError(e);
    }
  }
</script>

<div class="flex flex-row items-center w-[600px]">
  <Input size="sm" bind:value={filterText} onEnter={handleFilterChange} />
  {#if error}
    <Tooltip.Root portal="body">
      <Tooltip.Trigger>
        <InfoIcon class="text-red-500 ml-2" size="14px" strokeWidth={2} />
      </Tooltip.Trigger>
      <Tooltip.Content side="right">
        {error}
      </Tooltip.Content>
    </Tooltip.Root>
  {/if}
</div>
