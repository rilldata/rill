<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import type { MenuOption } from "./lib.ts";

  const functionNames = ["SUM", "AVG", "COUNT", "MIN", "MAX"] as const;
  const fields = ["Simple", "Advanced"];
  const simpleRegex = /^([a-zA-Z_][a-zA-Z0-9_]*)\(([a-zA-Z0-9_]+)\)$/;
  const label = "SQL expression";
  const id = "expression";

  export let columns: MenuOption[];
  export let numericColumns: MenuOption[];
  export let expression: string;
  export let name: string;
  export let editing: boolean;

  let viewingSimple = parseExpression(expression).simple;

  $: ({ parsedColumn, parsedFunction } = parseExpression(expression));

  $: finalColumns = parsedFunction === "COUNT" ? columns : numericColumns;

  $: nameMatchesExpression = !name || name === extractName(expression);

  function extractName(expression: string | undefined) {
    if (!expression) {
      return "";
    }
    const match = expression.match(simpleRegex);
    const column = match?.[2] ?? "";
    const functionName = match?.[1].toLocaleUpperCase() ?? functionNames[0];
    return createProperties(functionName, column).name;
  }

  function parseExpression(expression: string) {
    const match = expression.match(simpleRegex);
    const parsedColumn = match?.[2];
    const parsedFunction = match?.[1].toLocaleUpperCase() ?? functionNames[0];
    const simple =
      !expression.length ||
      (simpleRegex.test(expression) &&
        parsedColumn &&
        (parsedFunction === "COUNT" ||
          Boolean(numericColumns.find(({ value }) => value === parsedColumn))));

    return { parsedColumn, parsedFunction, simple };
  }

  function createProperties(functionName: string, column: string) {
    if (!functionName || !column) {
      return { name: "", expression: "" };
    }
    return {
      name: `${functionName.toLowerCase()}_of_${column}`,
      expression: `${functionName}(${column})`,
    };
  }
</script>

<div class="flex flex-col gap-y-1 h-fit">
  {#if label}
    <div class="label-wrapper">
      <label for={id} class="line-clamp-1">
        {label}
      </label>
    </div>
  {/if}

  <div class="rounded-sm option-wrapper flex h-6 text-sm w-fit mb-1">
    {#each fields as field (field)}
      <button
        on:click={() => {
          viewingSimple = field === "Simple";
          if (parsedFunction && parsedColumn) {
            const props = createProperties(parsedFunction, parsedColumn);
            if (nameMatchesExpression && !editing) name = props.name;
            expression = props.expression;
          }
        }}
        class="-ml-[1px] first-of-type:-ml-0 px-2 border border-gray-300 first-of-type:rounded-l-[2px] last-of-type:rounded-r-[2px]"
        class:selected={(viewingSimple && field === "Simple") ||
          (!viewingSimple && field === "Advanced")}
      >
        {field}
      </button>
    {/each}
  </div>

  <div class="flex gap-x-1.5 items-center">
    {#if viewingSimple}
      <Select
        ringFocus
        fontSize={14}
        id="vme-SQL expression"
        value={parsedFunction}
        options={functionNames.map((value) => ({ value, label: value }))}
        onChange={(newFunction) => {
          if (!parsedColumn) return;

          if (
            newFunction !== "COUNT" &&
            !numericColumns.find(({ value }) => value === parsedColumn)
          ) {
            expression = "";
            if (!editing) name = "";
          } else {
            const props = createProperties(newFunction, parsedColumn);

            if (props.name && nameMatchesExpression && !editing) {
              name = props.name;
            }
            if (props.expression) {
              expression = props.expression;
            }
          }
        }}
      />
      of
      {#key parsedFunction}
        <Select
          enableSearch
          ringFocus
          id="column"
          fontSize={14}
          full
          placeholder="Model column"
          value={parsedColumn}
          options={finalColumns ?? []}
          onChange={(newColumn) => {
            if (!parsedFunction) return;
            const props = createProperties(parsedFunction, newColumn);
            if (props.name && nameMatchesExpression && !editing) {
              name = props.name;
            }
            if (props.expression) {
              expression = props.expression;
            }
          }}
        />
      {/key}
    {:else}
      <Input
        textClass="text-sm"
        id="vme-SQL expression"
        full
        bind:value={expression}
        multiline
        fontFamily={`"Source Code Variable", monospace`}
      />
    {/if}
  </div>
</div>

<style lang="postcss">
  .label-wrapper {
    @apply flex items-center gap-x-1;
  }

  label {
    @apply text-sm font-medium text-gray-800;
  }

  .option-wrapper > .selected {
    @apply border-primary-500 z-50 text-primary-500;
  }
</style>
