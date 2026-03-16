<script lang="ts">
  import type { V1SecurityRule } from "@rilldata/web-common/runtime-client";
  import DescribeSection from "./DescribeSection.svelte";
  import DescribeRow from "./DescribeRow.svelte";

  export let rules: V1SecurityRule[] | undefined = undefined;
</script>

<DescribeSection title="Security Policy">
  {#if rules?.length}
    {#each rules as rule, i (i)}
      <div
        class="flex flex-col gap-y-1 {i > 0
          ? 'mt-1 pt-1 border-t border-border'
          : ''}"
      >
        {#if rule.access}
          <div class="flex flex-col gap-y-0.5">
            <span class="text-[11px] text-fg-secondary font-medium">Access</span
            >
            <DescribeRow
              label={rule.access.allow ? "Allow" : "Deny"}
              value={rule.access.conditionExpression || "all"}
            />
            {#if rule.access.exclusive}
              <span class="text-[11px] text-fg-muted pl-2">exclusive</span>
            {/if}
          </div>
        {/if}
        {#if rule.rowFilter}
          <div class="flex flex-col gap-y-0.5">
            <span class="text-[11px] text-fg-secondary font-medium"
              >Row filter</span
            >
            {#if rule.rowFilter.sql}
              <span class="text-[11px] text-fg-muted font-mono pl-2"
                >{rule.rowFilter.sql}</span
              >
            {/if}
            {#if rule.rowFilter.conditionExpression}
              <DescribeRow
                label="Condition"
                value={rule.rowFilter.conditionExpression}
              />
            {/if}
          </div>
        {/if}
        {#if rule.fieldAccess}
          <div class="flex flex-col gap-y-0.5">
            <span class="text-[11px] text-fg-secondary font-medium">
              Field {rule.fieldAccess.allow ? "include" : "exclude"}
            </span>
            {#if rule.fieldAccess.allFields}
              <span class="text-[11px] text-fg-muted pl-2">all fields</span>
            {:else if rule.fieldAccess.fields?.length}
              <span class="text-[11px] text-fg-muted font-mono pl-2">
                {rule.fieldAccess.fields.join(", ")}
              </span>
            {/if}
            {#if rule.fieldAccess.conditionExpression}
              <DescribeRow
                label="Condition"
                value={rule.fieldAccess.conditionExpression}
              />
            {/if}
            {#if rule.fieldAccess.exclusive}
              <span class="text-[11px] text-fg-muted pl-2">exclusive</span>
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  {:else}
    <span class="text-xs text-fg-muted">None defined</span>
  {/if}
</DescribeSection>
