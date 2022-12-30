<script lang="ts">
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import { clickOutside } from "@rilldata/web-common/lib/actions/click-outside";
  import { guidGenerator } from "@rilldata/web-common/lib/guid";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { slide } from "svelte/transition";
  import { flip } from "../row-flip";

  let duration = 200;

  export let items: any[];
  export let addItemText: string = undefined;

  let draftItems = items;
  /** refresh draftItems if items prop changes */
  $: draftItems = items;

  /** we utilize the  */
  let itemComponents = [];

  const dispatch = createEventDispatcher();

  function submitUpdatedItems() {
    dispatch("update-items", draftItems);
  }

  function newItem() {
    return {
      displayName: "",
      expression: "",
      id: guidGenerator(),
      visible: true,
    };
  }

  function addItem() {
    draftItems = [...draftItems, newItem()];
    submitUpdatedItems();
  }

  function deleteItem(id: string) {
    draftItems = [...draftItems.filter((item) => item.id !== id)];
    submitUpdatedItems();
  }

  function deactivateDragHandleMenus() {
    itemComponents.forEach((component) =>
      component?.deactivateDragHandleMenu()
    );
  }

  function moveUp(id: string) {
    let i = draftItems.findIndex((item) => item.id === id);
    if (i > 0 && selections.size < 2) {
      deactivateDragHandleMenus();

      const thisMeasure = { ...draftItems[i] };
      const otherMeasure = { ...draftItems[i - 1] };

      draftItems[i] = otherMeasure;
      draftItems[i - 1] = thisMeasure;
      // tell svelte to redraw
      draftItems = draftItems;
      submitUpdatedItems();
    }
  }

  async function moveDown(id: string) {
    let i = draftItems.findIndex((item) => item.id === id);
    if (i < draftItems.length - 1 && selections.size < 2) {
      deactivateDragHandleMenus();

      const thisMeasure = { ...draftItems[i] };
      const otherMeasure = { ...draftItems[i + 1] };

      draftItems[i] = otherMeasure;
      draftItems[i + 1] = thisMeasure;
      draftItems = draftItems;
      submitUpdatedItems();
    }
  }

  // here is where we listen to escape?
  let selections = new Set();

  function clearSelections() {
    selections = new Set();
  }

  function handleEdit(id: string) {
    return (event) => {
      let key = event.detail.key;
      let value = event.detail.value;
      console.log("hey", key, value);
      let item = draftItems.find((item) => item.id === id);
      item[key] = value;
      draftItems = draftItems;
      submitUpdatedItems();
    };
  }

  function handleSelect(id: string) {
    return (event) => {
      if (event.detail !== true) {
        // single select.
        selections = new Set([id]);
        return;
      }
      if (selections.has(id)) selections.delete(id);
      else selections.add(id);
      // tell svelte to update
      selections = selections;
    };
  }

  function handleToggleVisibility(id: string) {
    return () => {
      let item = draftItems.find((item) => item.id === id);
      item.visible = !item.visible;
      draftItems = draftItems;
      submitUpdatedItems();
    };
  }

  function handleDelete(event) {
    if (event.key === "Backspace" && event.shiftKey && selections.size > 0) {
      event.preventDefault();
      selections.forEach((id) => deleteItem(id));
      selections = new Set();
      draftItems = draftItems;
      submitUpdatedItems();
    }
  }

  function handleCancelSelection(event) {
    if (event.key === "Escape" && selections.size > 0) {
      selections = new Set();
      itemComponents.forEach((component) => component?.blurAllFields());
    }
  }

  function moveToBottom(id: string = undefined) {
    draftItems = [
      ...draftItems.filter((item) =>
        id ? item.id !== id : !selections.has(item.id)
      ),
      ...draftItems.filter((item) =>
        id ? item.id === id : selections.has(item.id)
      ),
    ];
    submitUpdatedItems();
  }

  function moveToTop(id: string = undefined) {
    draftItems = [
      ...draftItems.filter((item) =>
        id ? item.id === id : selections.has(item.id)
      ),
      ...draftItems.filter((item) =>
        id ? item.id !== id : !selections.has(item.id)
      ),
    ];
    submitUpdatedItems();
  }

  function handleMoveToOneSideOrOther(event) {
    if (selections.size > 0) {
      if (event.metaKey && event.key === "ArrowDown") {
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        deactivateDragHandleMenus();
        moveToBottom();
      }
      if (event.metaKey && event.key === "ArrowUp") {
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        deactivateDragHandleMenus();
        moveToTop();
      }
    }
  }

  function handleMoveUpOrDown(event) {
    if (selections.size === 1) {
      let selectionID = Array.from(selections)[0] as string;
      if (event.key === "ArrowDown" && event.shiftKey) {
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        moveDown(selectionID);
        itemComponents.forEach((component) => component?.blurAllFields());
      } else if (event.key === "ArrowUp" && event.shiftKey) {
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        moveUp(selectionID);
        itemComponents.forEach((component) => component?.blurAllFields());
      }
    }
  }

  function onKeydown(event) {
    handleCancelSelection(event);
    handleDelete(event);
    handleMoveUpOrDown(event);
    handleMoveToOneSideOrOther(event);
  }

  let dragY;
  let dragIndex;
  let every = 0;

  const END_POINT = guidGenerator();

  function onMousemove(event) {
    every += 1;
    if (isDragging && every % 5 === 0) {
      dragY = Math.max(0, event.clientY - containerTop);
      let indexMap = (dragY / containerSize) * draftItems.length;
      dragIndex = Math.min(draftItems.length - 1, ~~Math.round(indexMap));
      let candidate = activeItems[dragIndex].id;
      if (indexMap > draftItems.length - 1) {
        /** Set the candidate drag insertion point to be END_POINT; this will append
         * to the end of activeMeasures.
         */
        candidateDragInsertionPoint = END_POINT;
      } else if (candidate !== candidateDragInsertionPoint) {
        candidateDragInsertionPoint = candidate;
      }
    }
  }

  /** FIXME: how much of this can we deprecate */
  async function onMouseup() {
    draftItems = activeItems;
    isDragging = false;
    /** wait for the update before redrawing */
    dragID = undefined;
    candidateDragInsertionPoint = undefined;
    submitUpdatedItems();
  }

  /** drag id */

  let dragID: string = undefined;
  let isDragging = false;
  let candidateDragInsertionPoint = undefined;

  function handleDragHandleMousedown(id: string) {
    return (event) => {
      dragY = Math.max(0, event.detail.y - containerTop);
      isDragging = true;
      dragID = id;
    };
  }

  async function wait(ms) {
    await new Promise((resolve) => setTimeout(resolve, ms));
  }

  // three modes: editing, unselected, singleselect, multiselect.
  // editing only possible if not multiselect

  // click in first
  // unselected -> editing + singleselect
  // shift + click
  // unselected -> singleselect OR multiselect
  let mode: "unselected" | "select" | "multiselect";

  // update the mode according to the state.
  $: multiSelectActive = selections.size > 1;
  $: if (multiSelectActive) {
    mode = "multiselect";
  } else if (selections.size === 1) {
    mode = "select";
  } else {
    mode = "unselected";
  }

  $: if (mode === "multiselect")
    itemComponents.forEach((component) => component?.blurAllFields());

  let showError = false;

  let container;
  let containerSize = 0;
  let containerTop = 0;

  onMount(() => {
    let observer = new ResizeObserver(() => {
      let bb = container.getBoundingClientRect();
      containerSize = bb.height;
      containerTop = bb.top;
    });
    observer.observe(container);
  });
  /** just reorder the actual array! This will remove a LOT of complexity from the equation.
   * activeMeasures will be a temporarily
   *
   */
  // always reset activeMeasures if measures change.
  $: activeItems = draftItems;

  function handleDragging(dragID, candidateDragInsertionPoint) {
    if (
      dragID &&
      candidateDragInsertionPoint &&
      dragID !== candidateDragInsertionPoint
    ) {
      let original = activeItems.find((item) => item.id === dragID);
      let candidateActiveItems = [
        ...activeItems.filter((item) => item.id !== original.id),
      ];
      if (candidateDragInsertionPoint === END_POINT) {
        candidateActiveItems.push({ ...original });
      } else {
        let insertionIndex = candidateActiveItems.findIndex(
          (item) => item.id === candidateDragInsertionPoint
        );

        candidateActiveItems.splice(insertionIndex, 0, { ...original });
      }

      activeItems = candidateActiveItems;
    }
  }

  $: if (isDragging) {
    handleDragging(dragID, candidateDragInsertionPoint);
  }
</script>

<svelte:window
  on:keydown={onKeydown}
  on:mousemove={onMousemove}
  on:mouseup={onMouseup}
/>

<div use:clickOutside={[[], clearSelections]} bind:this={container}>
  <!-- iterate over the active items and slot in each item's state & callbacks -->
  {#each activeItems as item (item.id)}
    <div
      class="w-full"
      transition:slide|local={{ duration }}
      animate:flip={{
        duration,
        activeElement: dragID ? dragID === item.id : selections.has(item.id),
      }}
    >
      <slot
        {item}
        {mode}
        {isDragging}
        selected={selections.has(item.id)}
        edit={handleEdit(item.id)}
        moveUp={() => {
          moveUp(item.id);
        }}
        moveDown={() => {
          moveDown(item.id);
        }}
        moveToTop={() => {
          moveToTop(item.id);
        }}
        moveToBottom={() => {
          moveToBottom(item.id);
        }}
        deleteItem={async () => {
          await wait(150);
          deleteItem(item.id);
        }}
        select={handleSelect(item.id)}
        toggleVisibility={handleToggleVisibility(item.id)}
        dragHandleMousedown={handleDragHandleMousedown(item.id)}
      />
    </div>
  {/each}
  {#if addItemText}
    <button
      style:margin-left="20px"
      style:height="36px"
      class="flex items-center p-1 hover:bg-gray-100 w-full block gap-x-2 rounded"
      on:click={async () => {
        // measureComponents.forEach((measure) => measure?.blurAllFields());
        await tick();
        addItem();
      }}
    >
      <Add />
      {addItemText}
    </button>
  {/if}
</div>
