<script lang="ts">
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";

  const functionNames = ["SUM", "AVG", "COUNT", "MIN", "MAX"];
  const fields = ["Simple", "Advanced"];

  const simpleRegex = /^([a-zA-Z_][a-zA-Z0-9_]*)\(([a-zA-Z0-9_]+)\)$/;

  export let columnNames: string[];
  export let expression: string | undefined = undefined;
  export let name: string | undefined = undefined;
  export let editing: boolean;

  let label = "SQL expression";
  let id = "expression";
  let { column, functionName, simple } = isExistingExpressionSimple(
    expression ?? "",
  );

  $: nameMatchesExpression = !name || name === extractName(expression);

  function extractName(expression: string | undefined) {
    if (!expression) {
      return "";
    }
    const match = expression.match(simpleRegex);
    const column = match?.[2] ?? "";
    const functionName = match?.[1] ?? functionNames[0];
    return createProperties(functionName, column).name;
  }

  function isExistingExpressionSimple(expression: string) {
    const match = expression.match(simpleRegex);
    const column = match?.[2] ?? "";
    const functionName = match?.[1] ?? functionNames[0];

    const simple =
      !expression.length ||
      (simpleRegex.test(expression) &&
        columnNames.includes(column) &&
        functionNames.includes(functionName));

    return { column, functionName, simple };
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
          simple = field === "Simple";
          if (functionName && column) {
            const props = createProperties(functionName, column);
            if (nameMatchesExpression && !editing) name = props.name;
            expression = props.expression;
          }
        }}
        class="-ml-[1px] first-of-type:-ml-0 px-2 border border-gray-300 first-of-type:rounded-l-[2px] last-of-type:rounded-r-[2px]"
        class:selected={(simple && field === "Simple") ||
          (!simple && field === "Advanced")}
      >
        {field}
      </button>
    {/each}
  </div>

  <div class="flex gap-x-1.5 items-center">
    {#if simple}
      <Select
        ringFocus
        fontSize={14}
        id="vme-SQL expression"
        bind:value={functionName}
        options={functionNames.map((value) => ({ value, label: value })) ?? []}
        onChange={(newFunction) => {
          const props = createProperties(newFunction, column);

          if (props.name && nameMatchesExpression && !editing)
            name = props.name;
          if (props.expression) expression = props.expression;
        }}
      />
      of
      <Select
        ringFocus
        id="column"
        fontSize={14}
        full
        placeholder="Model column"
        bind:value={column}
        options={columnNames.map((value) => ({ value, label: value })) ?? []}
        onChange={(newColumn) => {
          const props = createProperties(functionName, newColumn);
          if (props.name && nameMatchesExpression && !editing)
            name = props.name;
          if (props.expression) expression = props.expression;
        }}
      />
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
