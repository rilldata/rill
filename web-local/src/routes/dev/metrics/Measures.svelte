<script lang="ts">
  import { clickOutside } from "@rilldata/web-common/components/actions/click-outside";
  import { guidGenerator } from "@rilldata/web-common/lib/util/guid";
  import { onMount, tick } from "svelte";
  import { slide } from "svelte/transition";
  import AddMeasure from "./AddMeasure.svelte";
  import Measure from "./dashboard-config/Measure.svelte";
  import { flip } from "./row-flip";

  let measureComponents = [];
  let duration = 200;

  interface Measure {
    displayName?: string;
    expression?: string;
    description?: string;
    id: string;
    visible: boolean;
  }

  let errorGUID = guidGenerator();

  let measures: Measure[] = [
    {
      displayName: "Total Records",
      expression: "count(*)",
      id: guidGenerator(),
      visible: true,
    },
    ...Array.from({ length: 0 }).map((_, i) => {
      return {
        displayName: "measure" + i,
        expression: `count(${i})`,
        id: guidGenerator(),
        visible: true,
      };
    }),
    {
      displayName: "Revenue",
      expression: "sum(revenue)",
      id: errorGUID,
      visible: true,
    },
    {
      displayName: "Distinct Users",
      expression: "count (distinct user_id)",
      id: guidGenerator(),
      visible: true,
    },
    {
      displayName: "Distinct Non-Bot Users",
      expression: "count (distinct sanitized_user_id)",
      id: guidGenerator(),
      visible: false,
    },
    {
      displayName: "Total Sales",
      expression: "sum(sales_price)",
      id: guidGenerator(),
      visible: true,
    },
  ];

  function newItem() {
    return {
      displayName: "",
      expression: "",
      id: guidGenerator(),
      visible: true,
    };
  }

  function addItem() {
    measures = [...measures, newItem()];
  }

  function deleteItem(id: string) {
    measures = [...measures.filter((measure) => measure.id !== id)];
  }

  function deactivateDragHandleMenus() {
    measureComponents.forEach((component) =>
      component?.deactivateDragHandleMenu()
    );
  }

  function moveUp(id: string) {
    let i = measures.findIndex((measure) => measure.id === id);
    if (i > 0 && selections.size < 2) {
      deactivateDragHandleMenus();

      const thisMeasure = { ...measures[i] };
      const otherMeasure = { ...measures[i - 1] };

      measures[i] = otherMeasure;
      measures[i - 1] = thisMeasure;
      measures = measures;
    }
  }

  async function moveDown(id: string) {
    let i = measures.findIndex((measure) => measure.id === id);
    if (i < measures.length - 1 && selections.size < 2) {
      deactivateDragHandleMenus();

      const thisMeasure = { ...measures[i] };
      const otherMeasure = { ...measures[i + 1] };

      measures[i] = otherMeasure;
      measures[i + 1] = thisMeasure;
      measures = measures;
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
      let measure = measures.find((measure) => measure.id === id);
      measure[key] = value;
      measures = measures;
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
      let measure = measures.find((measure) => measure.id === id);
      measure.visible = !measure.visible;
      measures = measures;
    };
  }

  function handleDelete(event) {
    if (event.key === "Backspace" && event.shiftKey && selections.size > 0) {
      event.preventDefault();
      selections.forEach((id) => deleteItem(id));
      selections = new Set();
      measures = measures;
    }
  }

  function handleCancelSelection(event) {
    if (event.key === "Escape" && selections.size > 0) {
      selections = new Set();
      measureComponents.forEach((component) => component?.blurAllFields());
    }
  }

  function moveToBottom(id: string = undefined) {
    measures = [
      ...measures.filter((measure) =>
        id ? measure.id !== id : !selections.has(measure.id)
      ),
      ...measures.filter((measure) =>
        id ? measure.id === id : selections.has(measure.id)
      ),
    ];
  }

  function moveToTop(id: string = undefined) {
    measures = [
      ...measures.filter((measure) =>
        id ? measure.id === id : selections.has(measure.id)
      ),
      ...measures.filter((measure) =>
        id ? measure.id !== id : !selections.has(measure.id)
      ),
    ];
  }

  function handleMoveToOneSideOrOther(event) {
    if (selections.size > 0) {
      if (event.metaKey && event.key === "ArrowDown") {
        deactivateDragHandleMenus();
        moveToBottom();
      }
      if (event.metaKey && event.key === "ArrowUp") {
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
        moveDown(selectionID);
        measureComponents.forEach((component) => component?.blurAllFields());
      } else if (event.key === "ArrowUp" && event.shiftKey) {
        event.preventDefault();
        moveUp(selectionID);
        measureComponents.forEach((component) => component?.blurAllFields());
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
      let indexMap = (dragY / containerSize) * measures.length;
      dragIndex = Math.min(measures.length - 1, ~~Math.round(indexMap));
      let candidate = activeMeasures[dragIndex].id;
      if (indexMap > measures.length - 1) {
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
    measures = activeMeasures;
    isDragging = false;
    /** wait for the update before redrawing */
    dragID = undefined;
    candidateDragInsertionPoint = undefined;
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
    measureComponents.forEach((component) => component?.blurAllFields());

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
  $: activeMeasures = measures;

  function handleDragging(dragID, candidateDragInsertionPoint) {
    if (
      dragID &&
      candidateDragInsertionPoint &&
      dragID !== candidateDragInsertionPoint
    ) {
      let original = activeMeasures.find((measure) => measure.id === dragID);
      let candidateActiveMeasures = [
        ...activeMeasures.filter((measure) => measure.id !== original.id),
      ];
      if (candidateDragInsertionPoint === END_POINT) {
        candidateActiveMeasures.push({ ...original });
      } else {
        let insertionIndex = candidateActiveMeasures.findIndex(
          (measure) => measure.id === candidateDragInsertionPoint
        );

        candidateActiveMeasures.splice(insertionIndex, 0, { ...original });
      }

      activeMeasures = candidateActiveMeasures;
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

<input type="checkbox" bind:checked={showError} />

<h1>Measures</h1>

<div use:clickOutside={[[], clearSelections]} bind:this={container}>
  {#each activeMeasures as measure, i (measure.id)}
    <div
      class="w-full"
      transition:slide|local={{ duration }}
      animate:flip={{
        duration,
        activeElement: dragID
          ? dragID === measure.id
          : selections.has(measure.id),
      }}
    >
      <Measure
        error={showError && measure.id === errorGUID
          ? "This is what an error would look like"
          : undefined}
        bind:this={measureComponents[i]}
        visible={measure.visible}
        expression={measure.expression}
        displayName={measure.displayName}
        description={measure.description}
        selected={selections.has(measure.id)}
        {mode}
        {isDragging}
        on:draghandle-mousedown={handleDragHandleMousedown(measure.id)}
        on:select={handleSelect(measure.id)}
        on:toggle-visibility={handleToggleVisibility(measure.id)}
        on:edit={handleEdit(measure.id)}
        on:delete={async () => {
          await wait(150);
          deleteItem(measure.id);
        }}
        on:move-up={() => {
          moveUp(measure.id);
        }}
        on:move-down={() => {
          moveDown(measure.id);
        }}
        on:move-to-top={() => {
          moveToTop(measure.id);
        }}
        on:move-to-bottom={() => {
          moveToBottom(measure.id);
        }}
      />
    </div>
  {/each}
</div>
<AddMeasure
  on:click={async () => {
    measureComponents.forEach((measure) => measure?.blurAllFields());
    await tick();
    addItem();
  }}
/>
