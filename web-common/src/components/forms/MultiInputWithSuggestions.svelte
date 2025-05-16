<script lang="ts">
  import { createEventDispatcher, onMount, tick } from "svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";

  type Suggestion = {
    value: string;
    label: string;
    type?: string;
    photoUrl?: string;
  };

  export let suggestions: Suggestion[] = [];
  export let values: string[] = [""];
  export let placeholder: string = "Search or add values, separated by commas";
  export let contentClassName: string = "relative";
  export let errors: Record<string | number, string[]> | undefined = undefined;
  export let singular: string = "item";
  export let plural: string = "items";
  export let preventFocus: boolean = false;
  export let id: string = "multi-input-with-suggestions";

  const dispatch = createEventDispatcher();

  let inputValue = "";
  let showSuggestions = false;
  let filteredSuggestions: Suggestion[] = suggestions.slice(0, 10);
  let multiInputEl: HTMLElement;
  let activeInput: HTMLInputElement | null = null;

  // Initialize suggestions and listeners on mount
  onMount(() => {
    // Initial suggestions
    filteredSuggestions = suggestions.slice(0, 10);

    // We need to add event listeners directly to handle MultiInput's complex structure
    setTimeout(() => {
      setupInputListeners();
    }, 0);
  });

  // Called on mount and whenever values change (which might mean new inputs)
  function setupInputListeners() {
    if (!multiInputEl) return;

    // Find all input elements inside MultiInput
    const inputs = multiInputEl.querySelectorAll("input");
    inputs.forEach((input) => {
      // Add direct event listeners to catch typing
      input.addEventListener("input", (e) => {
        const target = e.target as HTMLInputElement;
        inputValue = target.value || "";
        activeInput = target;
        updateFilteredSuggestions();
        if (inputValue.trim() && filteredSuggestions.length > 0) {
          showSuggestions = true;
        }
      });

      input.addEventListener("focus", () => {
        activeInput = input;
        showSuggestions = true;
        if (suggestions.length > 0) {
          filteredSuggestions = suggestions.slice(0, 10);
        }
      });
    });
  }

  // Re-setup listeners when values change (new inputs might be created)
  $: if (values) {
    tick().then(() => {
      setupInputListeners();
    });
  }

  // Update suggestions when they change
  $: if (suggestions) {
    updateFilteredSuggestions();
  }

  function handleInputFocus() {
    showSuggestions = true;
    if (suggestions.length > 0) {
      filteredSuggestions = suggestions.slice(0, 10);
    }
  }

  function handleInputBlur() {
    setTimeout(() => {
      showSuggestions = false;
    }, 200);
  }

  // This is still used by the MultiInput events, but we mainly rely on direct DOM events
  function handleInputChange(e) {
    if (e.detail?.value !== undefined) {
      inputValue = e.detail.value;
      updateFilteredSuggestions();
      if (inputValue.trim() && filteredSuggestions.length > 0) {
        showSuggestions = true;
      }
    }
  }

  function handleSuggestionClick(suggestion: Suggestion) {
    let index = 0;

    // Try to find which input is active
    if (activeInput && multiInputEl) {
      const inputs = Array.from(multiInputEl.querySelectorAll("input"));
      index = inputs.indexOf(activeInput);
      if (index === -1) index = values.length - 1;
    } else {
      // Fallback to finding empty input
      index = values.findIndex((v) => !v.trim());
      if (index === -1) index = values.length - 1;
    }

    // Update values with suggestion
    const newValues = [...values];
    newValues[index] = suggestion.value;

    // Add empty input at the end if needed
    if (index === newValues.length - 1) {
      newValues.push("");
    }

    // Update values and dispatch change event
    values = newValues;
    dispatch("change", { values: newValues });

    // Reset input
    inputValue = "";
    showSuggestions = false;

    // Re-setup listeners after values change
    tick().then(() => {
      setupInputListeners();
    });
  }

  function updateFilteredSuggestions() {
    if (!inputValue || !inputValue.trim()) {
      // Show all suggestions when input is empty
      filteredSuggestions = suggestions.slice(0, 10);
    } else {
      // Filter suggestions based on input
      filteredSuggestions = suggestions
        .filter(
          (s) =>
            s.label.toLowerCase().includes(inputValue.toLowerCase()) ||
            s.value.toLowerCase().includes(inputValue.toLowerCase()),
        )
        .slice(0, 10);
    }
  }
</script>

<div class="relative w-full" bind:this={multiInputEl}>
  <MultiInput
    {id}
    {placeholder}
    {contentClassName}
    bind:values
    {errors}
    {singular}
    {plural}
    {preventFocus}
    on:focus={handleInputFocus}
    on:blur={handleInputBlur}
    on:input={handleInputChange}
    on:change={handleInputChange}
  >
    <slot name="within-input" slot="within-input"></slot>
    <slot name="beside-input" slot="beside-input" let:hasSomeValue>
      <slot name="action-button" {hasSomeValue}></slot>
    </slot>
  </MultiInput>

  <!-- Show dropdown only when we have suggestions and showSuggestions is true -->
  {#if showSuggestions && filteredSuggestions.length > 0}
    <div
      class="absolute z-10 mt-1 w-full max-h-[208px] overflow-y-auto rounded-sm border border-gray-200 bg-white shadow-md"
    >
      {#each filteredSuggestions as suggestion}
        <button
          type="button"
          class="flex items-center p-2 w-full text-left hover:bg-gray-100 cursor-pointer"
          on:click={() => handleSuggestionClick(suggestion)}
          on:keydown={(e) =>
            e.key === "Enter" && handleSuggestionClick(suggestion)}
        >
          {#if suggestion.type === "user"}
            <div class="flex items-center gap-2">
              <div
                class="h-6 w-6 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden"
              >
                {#if suggestion.photoUrl}
                  <img
                    src={suggestion.photoUrl}
                    alt={suggestion.label}
                    class="h-full w-full object-cover"
                  />
                {:else}
                  <span class="text-xs text-gray-700"
                    >{suggestion.label[0]?.toUpperCase()}</span
                  >
                {/if}
              </div>
              <div class="flex flex-col">
                <span class="text-sm">{suggestion.label}</span>
                {#if suggestion.value !== suggestion.label}
                  <span class="text-xs text-gray-500">{suggestion.value}</span>
                {/if}
              </div>
            </div>
          {:else}
            <div class="flex items-center gap-2">
              <div
                class="h-6 w-6 rounded-sm bg-primary-600 flex items-center justify-center"
              >
                <span class="text-xs text-white"
                  >{suggestion.label[0]?.toUpperCase()}</span
                >
              </div>
              <span class="text-sm">{suggestion.label}</span>
            </div>
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>
