<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import CriteriaForm from "@rilldata/web-common/features/alerts/criteria-tab/CriteriaForm.svelte";
  import { createBinaryExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { Trash2Icon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";

  export let expr: V1Expression;

  const dispatch = createEventDispatcher();

  function handleDeleteCriteria(index: number) {
    expr.cond?.exprs?.splice(index, 1);
    dispatch("update", expr);
  }

  function handleAddCriteria() {
    expr.cond?.exprs?.push(
      createBinaryExpression("", V1Operation.OPERATION_UNSPECIFIED, 0),
    );
    dispatch("update", expr);
  }

  function handleUpdateCriteria(index: number, subExpr: V1Expression) {
    expr.cond?.exprs?.splice(index, 1, subExpr);
    dispatch("update", expr);
  }
</script>

{#if expr.cond?.exprs}
  <div class="flex flex-col gap-2">
    {#each expr.cond.exprs as subExpr, i}
      <div class="flex flex-col gap-2">
        <div class="flex flex-row items-center">
          <div class="w-full text-lg">{i + 1}</div>
          <button class="mr-2" on:click={() => handleDeleteCriteria(i)}>
            <Trash2Icon size="16px" />
          </button>
        </div>
        <CriteriaForm
          expr={subExpr}
          on:update={(e) => handleUpdateCriteria(i, e.detail)}
        />
      </div>
    {/each}
    <Button dashed on:click={handleAddCriteria}>+ Add Criteria</Button>
  </div>
{/if}
