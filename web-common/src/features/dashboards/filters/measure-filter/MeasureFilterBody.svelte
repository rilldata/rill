<script lang="ts">
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import SelectMenu from "@rilldata/web-common/components/menu/compositions/SelectMenu.svelte";
  import type { SelectMenuItem } from "@rilldata/web-common/components/menu/types";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";

  export let expr: V1Expression | undefined;

  const options: Array<SelectMenuItem> = [
    {
      key: V1Operation.OPERATION_EQ,
      main: "=",
    },
    {
      key: V1Operation.OPERATION_NEQ,
      main: "!=",
    },
    {
      key: V1Operation.OPERATION_LT,
      main: "<",
    },
    {
      key: V1Operation.OPERATION_LTE,
      main: "<=",
    },
    {
      key: V1Operation.OPERATION_GT,
      main: ">",
    },
    {
      key: V1Operation.OPERATION_GTE,
      main: ">=",
    },
    // TODO
    // {
    //   key: "b",
    //   main: "Between",
    // },
    // {
    //   key: "nb",
    //   main: "Not Between",
    // },
  ];
  $: selection = options.find((o) => o.key === expr.cond?.op);
  $: val1 = expr?.cond?.exprs?.[1]?.val as string;
  let err1: string;
  $: val2 = expr?.cond?.exprs?.[2]?.val as string;
  let err2: string;

  $: console.log(selection, val1, err1, val2, err2);
</script>

<SelectMenu ariaLabel="Select a filter for measure" bind:selection {options} />
{#if selection}
  <InputV2 bind:value={val1} bind:error={err1} />
  {#if selection.key === "b" || selection.key === "nb"}
    <InputV2 bind:value={val2} bind:error={err2} />
  {/if}
{/if}
