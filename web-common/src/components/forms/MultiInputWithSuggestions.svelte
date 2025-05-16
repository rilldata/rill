<script lang="ts">
  import { createEventDispatcher, onMount, tick, onDestroy } from "svelte";
  import MultiInput from "@rilldata/web-common/components/forms/MultiInput.svelte";
  import AvatarListItem from "@rilldata/web-admin/features/organizations/users/AvatarListItem.svelte";
  import { PillManager } from "./pill-manager";

  type Suggestion = {
    value: string;
    label: string;
    type?: string;
    photoUrl?: string;
  };

  type PillItem = {
    id: string;
    value: string;
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
  const pillManager = new PillManager(values, (newValues) => {
    values = newValues;
    dispatch("change", { values: newValues });
  });

  let inputValue = "";
  let showSuggestions = false;
  let filteredSuggestions: Suggestion[] = suggestions.slice(0, 10);
  let multiInputEl: HTMLElement;
  let activeInputId: string | null = null;

  // Store references to event handlers so we can remove them later
  let inputEventHandlers = new Map();
  let mousedownEventHandlers = new Map();

  // Initialize listeners on mount
  onMount(() => {
    // Initial suggestions
    filteredSuggestions = suggestions.slice(0, 10);

    // Setup input element event listeners
    setTimeout(() => {
      setupInputListeners();
    }, 0);
  });

  // Clean up all event listeners when component is destroyed
  onDestroy(() => {
    if (inputEventHandlers) {
      inputEventHandlers.forEach((handler, input) => {
        input.removeEventListener("input", handler);
      });
    }

    if (mousedownEventHandlers) {
      mousedownEventHandlers.forEach((handler, input) => {
        input.removeEventListener("mousedown", handler);
      });
    }

    inputEventHandlers.clear();
    mousedownEventHandlers.clear();
  });

  // Add direct event listeners to each input
  function setupInputListeners() {
    if (!multiInputEl) return;

    // First, remove any existing event listeners to prevent duplicates
    inputEventHandlers.forEach((handler, input) => {
      input.removeEventListener("input", handler);
    });
    mousedownEventHandlers.forEach((handler, input) => {
      input.removeEventListener("mousedown", handler);
    });

    // Clear the maps
    inputEventHandlers.clear();
    mousedownEventHandlers.clear();

    // Find all input elements inside MultiInput and map them to their IDs
    const inputs = multiInputEl.querySelectorAll("input");
    inputs.forEach((input) => {
      const inputId = input.getAttribute("data-pill-id");

      // Create and store input event handler
      const inputHandler = (e) => {
        const target = e.target as HTMLInputElement;
        inputValue = target.value || "";
        activeInputId = inputId;
        updateFilteredSuggestions();

        // Update the pill value directly
        if (inputId) {
          pillManager.updatePillValue(inputId, inputValue);
        }

        if (filteredSuggestions.length > 0) {
          showSuggestions = true;
        }
      };
      inputEventHandlers.set(input, inputHandler);
      input.addEventListener("input", inputHandler);

      // Create and store mousedown event handler
      const mousedownHandler = (e) => {
        activeInputId = inputId;
        showSuggestions = !showSuggestions;
        if (showSuggestions) {
          updateFilteredSuggestions();
        }
      };
      mousedownEventHandlers.set(input, mousedownHandler);
      input.addEventListener("mousedown", mousedownHandler);
    });
  }

  // Re-setup listeners when values change
  $: if (values) {
    tick().then(() => {
      setupInputListeners();
    });
  }

  // Keep pillIds in sync with pillItems
  $: pillIds = pillManager.getPills().map((pill) => pill.id);

  // Update suggestions when they change
  $: if (suggestions) {
    updateFilteredSuggestions();
  }

  function handleInputFocus() {
    // Don't automatically show suggestions on focus
    // as we're now controlling this with clicks
  }

  function handleInputBlur() {
    // Small delay to allow clicking suggestions
    setTimeout(() => {
      showSuggestions = false;
    }, 200);
  }

  function handleInputChange(e) {
    if (e.detail?.value !== undefined) {
      inputValue = e.detail.value;
      updateFilteredSuggestions();

      // Only show dropdown when typing (not when clicking to toggle)
      if (inputValue.trim() && filteredSuggestions.length > 0) {
        showSuggestions = true;
      }

      // Find which pill to update based on activeInputId
      if (activeInputId) {
        pillManager.updatePillValue(activeInputId, inputValue);
      }
    }

    // Also handle pillItems changes
    if (e.detail?.values) {
      pillManager.setValues(e.detail.values);
    }
  }

  function handleSuggestionClick(suggestion: Suggestion) {
    // Find which pill to update based on activeInputId
    let pillIndex = -1;

    if (activeInputId) {
      const pills = pillManager.getPills();
      pillIndex = pills.findIndex((pill) => pill.id === activeInputId);
    }

    // If no active input ID or not found, find first empty pill
    if (pillIndex === -1) {
      const pills = pillManager.getPills();
      pillIndex = pills.findIndex((pill) => !pill.value.trim());
      if (pillIndex === -1) {
        // If no empty pill, use the last one
        pillIndex = pills.length - 1;
      }
    }

    // Update the pill value
    const pills = pillManager.getPills();
    if (pillIndex !== -1) {
      pillManager.updatePillValue(pills[pillIndex].id, suggestion.value);
    }

    // Reset state
    inputValue = "";
    showSuggestions = false;
    activeInputId = null;

    // Focus the empty input after DOM update
    setTimeout(() => {
      if (multiInputEl) {
        // Get all inputs and focus the last one (which should be empty)
        const inputs = Array.from(multiInputEl.querySelectorAll("input"));
        const lastInput = inputs[inputs.length - 1];
        if (lastInput) {
          lastInput.focus();
          activeInputId = lastInput.getAttribute("data-pill-id");
        }
      }

      // Re-setup listeners
      setupInputListeners();
    }, 10);
  }

  function updateFilteredSuggestions() {
    // Get currently selected values
    const selectedValues = pillManager.getValues().filter(Boolean);

    if (!inputValue || !inputValue.trim()) {
      // Show all suggestions when input is empty, but exclude already selected values
      filteredSuggestions = suggestions
        .filter((s) => !selectedValues.includes(s.value))
        .slice(0, 10);
    } else {
      // Filter suggestions based on input and exclude already selected values
      filteredSuggestions = suggestions
        .filter(
          (s) =>
            (s.label.toLowerCase().includes(inputValue.toLowerCase()) ||
              s.value.toLowerCase().includes(inputValue.toLowerCase())) &&
            !selectedValues.includes(s.value),
        )
        .slice(0, 10);
    }
  }
</script>

<div class="relative w-full" bind:this={multiInputEl}>
  <div class="custom-multi-input">
    <MultiInput
      {id}
      {placeholder}
      {contentClassName}
      values={pillManager.getValues()}
      {pillIds}
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
  </div>

  <!-- Show dropdown only when we have suggestions and showSuggestions is true -->
  {#if showSuggestions && filteredSuggestions.length > 0}
    <div
      class="absolute z-10 mt-1 w-full max-h-[208px] overflow-y-auto rounded-sm border border-gray-200 bg-white shadow-md"
    >
      {#each filteredSuggestions as suggestion}
        <button
          type="button"
          class="w-full text-left hover:bg-gray-100 cursor-pointer"
          on:click={() => handleSuggestionClick(suggestion)}
          on:keydown={(e) =>
            e.key === "Enter" && handleSuggestionClick(suggestion)}
        >
          {#if suggestion.type === "user"}
            <AvatarListItem
              name={suggestion.label}
              email={suggestion.value !== suggestion.label
                ? suggestion.value
                : null}
              photoUrl={suggestion.photoUrl}
              shape="circle"
            />
          {:else}
            <!-- TODO: add count -->
            <AvatarListItem name={suggestion.label} shape="square" />
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>
