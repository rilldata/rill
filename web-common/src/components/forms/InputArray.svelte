<script lang="ts">
  import type { createForm } from "svelte-forms-lib";
  import { slide } from "svelte/transition";
  import { Button, IconButton } from "../button";
  import Add from "../icons/Add.svelte";
  import InfoCircle from "../icons/InfoCircle.svelte";
  import Trash from "../icons/Trash.svelte";
  import Tooltip from "../tooltip/Tooltip.svelte";
  import TooltipContent from "../tooltip/TooltipContent.svelte";

  export let id: string;
  export let label = "";
  export let description = "";
  // The accessorKey is necessary due to the way svelte-forms-lib works with arrays.
  // See: https://svelte-forms-lib-sapper-docs.vercel.app/array
  export let accessorKey: string;
  export let placeholder = "";
  export let hint = "";
  export let addItemLabel = "Add item";

  export let formState: ReturnType<typeof createForm<Record<string, any>>>;
  const { form, errors } = formState;

  $: values = $form[id] as Record<string, string>[];
  // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
  // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
  $: errs = ($errors[id] as unknown as Record<string, string>[]) ?? [];

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      event.preventDefault();
    }
  }

  function handleAddItem() {
    $form[id] = $form[id].concat({ email: "" });
    errs = errs.concat({ email: "" });

    // Focus on the new input element
    setTimeout(() => {
      const input = document.getElementById(
        `${id}.${$form[id].length - 1}.${accessorKey}`,
      );
      input?.focus();
    }, 0);
  }

  function handleRemove(index: number) {
    $form[id] = $form[id].filter((_, i) => i !== index);
    errs = errs.filter((r, i) => i !== index);
  }
</script>

<div class="flex flex-col gap-y-2.5">
  {#if label}
    <div class="flex items-center gap-x-1">
      <label for={id} class="text-gray-800 text-sm font-medium">{label}</label>
      {#if hint}
        <Tooltip location="right" alignment="middle" distance={8}>
          <div class="text-gray-500" style="transform:translateY(-.5px)">
            <InfoCircle size="13px" />
          </div>
          <TooltipContent maxWidth="400px" slot="tooltip-content">
            {hint}
          </TooltipContent>
        </Tooltip>
      {/if}
      {#if description}
        <div class="text-sm text-slate-600">{description}</div>
      {/if}
    </div>
  {/if}
  <div
    class="flex flex-col gap-y-4 max-h-[200px] pl-1 pr-4 py-1 overflow-y-auto"
  >
    {#each values as _, i}
      <div class="flex flex-col gap-y-2">
        <div class="flex gap-x-2 items-center">
          <input
            bind:value={values[i][accessorKey]}
            id="{id}.{i}.{accessorKey}"
            autocomplete="off"
            {placeholder}
            class="bg-white rounded-sm border border-gray-300 px-3 py-[5px] h-8 cursor-pointer focus:outline-primary-500 w-full text-xs {errors[
              i
            ]?.accessorKey && 'border-red-500'}"
            on:keydown={handleKeyDown}
          />
          <IconButton on:click={() => handleRemove(i)}>
            <Trash size="16px" className="text-gray-500 cursor-pointer" />
          </IconButton>
        </div>
        {#if errs[i]?.[accessorKey]}
          <div in:slide={{ duration: 200 }} class="text-red-500 text-sm py-px">
            {errs[i][accessorKey]}
          </div>
        {/if}
      </div>
    {/each}
    <Button dashed on:click={handleAddItem} type="secondary">
      <div class="flex gap-x-2">
        <Add className="text-gray-700" />
        {addItemLabel}
      </div>
    </Button>
  </div>
</div>
