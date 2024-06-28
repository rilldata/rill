<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";

  /**
   * Input that allows to enter multiple items but appears within a single input box.
   * This is a more advanced version of InputArray.svelte
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { AlertCircleIcon, Trash, XIcon } from "lucide-svelte";
  import type { createForm } from "svelte-forms-lib";

  export let id: string;
  export let label = "";
  export let description = "";
  export let accessorKey: string;
  export let hint = "";

  export let formState: ReturnType<typeof createForm<Record<string, any>>>;
  const { form, errors, validateField } = formState;
  $: values = $form[id] as Record<string, string>[];
  // There's a bug in how `svelte-forms-lib` types the `$errors` store for arrays.
  // See: https://github.com/tjinauyeung/svelte-forms-lib/issues/154#issuecomment-1087331250
  $: errs = ($errors[id] as unknown as Record<string, string>[]) ?? [];

  let editingInput = "";
  let focused = false;
  function handleKeyDown(event: KeyboardEvent) {
    if (event.key !== "Enter") return;

    event.preventDefault();
    $form[id] = $form[id].concat({ email: editingInput });
    errs = errs.concat({ email: "" });
    editingInput = "";
    validateField(`${id}.${$form[id].length - 1}.${accessorKey}`);
  }

  function handleRemove(index: number) {
    $form[id] = $form[id].filter((_, i) => i !== index);
    errs = errs.filter?.((r, i) => i !== index);
  }

  $: hasSomeErrors = errs.some((e) => !!e[accessorKey]);
</script>

<div class="flex flex-col">
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
    class="flex flex-wrap gap-2 bg-white rounded-sm border border-gray-300 px-3 py-[5px] w-full"
    class:outline={focused || hasSomeErrors}
    class:outline-red-500={hasSomeErrors}
    class:outline-primary-500={focused && !hasSomeErrors}
  >
    {#each values as _, i}
      {@const hasError = !!errs[i]?.[accessorKey]}
      <div
        class="flex items-center text-gray-600 text-sm rounded-2xl border border-gray-300 bg-gray-100 pl-2 pr-1 max-w-full"
      >
        {#if hasError}
          <Tooltip distance={8}>
            <div class="w-fit h-5 overflow-hidden text-ellipsis text-red-500">
              {values[i][accessorKey]}
            </div>
            <TooltipContent maxWidth="400px" slot="tooltip-content">
              {errs[i][accessorKey]}
            </TooltipContent>
          </Tooltip>
        {:else}
          <div class="w-fit h-5 overflow-hidden text-ellipsis">
            {values[i][accessorKey]}
          </div>
        {/if}
        <IconButton disableHover on:click={() => handleRemove(i)}>
          <XIcon size="16px" className="text-gray-500 cursor-pointer" />
        </IconButton>
      </div>
    {/each}
    <input
      bind:value={editingInput}
      on:keydown={handleKeyDown}
      autocomplete="off"
      id="{id}.{values.length}.{accessorKey}"
      class="focus:outline-white group-hover:text-red-500 text-sm w-full"
      on:focusin={() => (focused = true)}
      on:focusout={() => (focused = false)}
    />
  </div>
</div>
