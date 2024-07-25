<script lang="ts">
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import { measureChipColors as colors } from "@rilldata/web-common/components/chip/chip-types";
  import { Mouse } from "lucide-svelte";
  import EditControls from "./EditControls.svelte";
  import { GripVerticalIcon, GripVertical } from "lucide-svelte";
  import Checkbox from "./Checkbox.svelte";
  const ROW_HEIGHT = 40;

  export let measure: MetricsViewSpecMeasureV2;
  export let reorderList: (initIndex: number, newIndex: number) => void;
  export let i: number;

  //   export let swapRows: (i: number) => void;

  let row: HTMLTableRowElement;
  let initialY = 0;
  let dragging = false;
  let rowDelta = 0;
  let swaps = 0;
  let y = 0;

  function handleClick(e: MouseEvent) {
    if (e.button !== 0) return;
    initialY = e.clientY;
    dragging = true;
    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);

    // const cloned = row.cloneNode(true) as HTMLTableRowElement;
    // cloned.style.position = "absolute";
    // cloned.style.width = "100%";
    // row.parentElement?.appendChild(cloned);d
  }

  function swapRows(down: boolean) {
    console.log({ down });
    const swapRow = (
      down ? row.nextElementSibling : row.previousElementSibling
    ) as HTMLTableRowElement;

    // console.log({ swapRow });
    initialY = swapRow.getBoundingClientRect().y + 20.5;

    // console.log({ initialY });
    // console.log(tbody.children[i + swaps + 2]);
    row.insertAdjacentElement(down ? "beforebegin" : "afterend", swapRow);
    // row.style.transform = `translateY(${down ? "-" : ""}20.5px)`;
    y = down ? -20.5 : 20.5;
    // tbody.insertBefore(nextRow, row);
  }

  function handleMouseMove(event: MouseEvent) {
    event.preventDefault();
    const { clientY } = event;
    y = clientY - initialY;
    console.log({ y, clientY, initialY });

    // rowDelta = Math.floor((y + 20) / 40);

    // console.log({ y });

    // console.log(y);
    // console.log({ rowDelta });

    if (y > 20.5 || y < -20.5) {
      console.log({ y });
      swapRows(y > 0);
      swaps += y > 0 ? -1 : 1;
      //   console.log({ initialY });

      //   y = -20;
    }
  }

  function handleMouseUp() {
    window.removeEventListener("mousemove", handleMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);

    const newIndex = i + swaps;
    if (newIndex !== i) {
      reorderList(i, newIndex);
    }

    y = 0;
    rowDelta = 0;
    dragging = false;
    swaps = 0;
  }
  let hovered = false;
</script>

<tr
  id={measure.name}
  class:dragging
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  bind:this={row}
  style:transform="translateY({y}px)"
  class="relative"
>
  <td class="!pl-0">
    <div class="h-10 pl-1 gap-x-0.5 flex items-center w-14">
      <button on:mousedown={handleClick}>
        <GripVertical size="14px" />
      </button>
      <Checkbox />
    </div>
  </td>
  <td>
    <Chip
      {...colors}
      extraRounded={false}
      extraPadding={false}
      label={measure.name}
      outline
    >
      <span slot="body" class="font-bold truncate">{measure.label}</span>
    </Chip>
  </td>
  <td class="expression">{measure.expression}</td>
  <td class="capitalize">{measure.formatPreset || measure.formatD3 || "-"}</td>
  <td>{measure.description || "-"}</td>

  {#if hovered}
    <EditControls />
  {/if}
</tr>

<style lang="postcss">
  tr {
    @apply border-b bg-background;
    height: 40px;
  }

  tr:hover {
    @apply bg-gray-50;
  }

  td {
    @apply pl-4 truncate;
  }

  .expression {
    font-family: "Source Code Variable", monospace;
    text-transform: uppercase;
  }

  .dragging {
    /* position: absolute; */
    width: 100%;
    /* display: block; */
    /* display: table-row; */
    /* border: 1px solid #f1f1f1; */
    z-index: 50;
    cursor: grabbing;
    /* -webkit-box-shadow: 2px 2px 3px 0px rgba(0, 0, 0, 0.05); */
    /* box-shadow: 2px 2px 3px 0px rgba(0, 0, 0, 0.05); */

    opacity: 1;
  }

  .dragging td {
    z-index: 50;
  }

  .container input {
    position: absolute;
    opacity: 0;
    cursor: pointer;
    height: 0;
    width: 0;
  }

  .checkmark {
    /* position: absolute;
    top: 0;
    left: 0; */
    height: 16px;
    width: 16px;
    @apply rounded-sm border border-gray-300;
    @apply bg-gray-50;
  }

  .container {
    @apply bg-green-400;
    @apply size-4;
    cursor: pointer;
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
  }

  .checkbox {
    @apply size-4 bg-gray-50 border border-gray-300;
  }

  /* On mouse-over, add a grey background color */
  .container:hover input ~ .checkmark {
    background-color: #ccc;
  }

  /* When the checkbox is checked, add a blue background */
  .container input:checked ~ .checkmark {
    background-color: #2196f3;
  }

  /* Create the checkmark/indicator (hidden when not checked) */
  .checkmark:after {
    content: "";
    position: absolute;
    display: none;
  }

  /* Show the checkmark when checked */
  .container input:checked ~ .checkmark:after {
    display: block;
  }

  /* Style the checkmark/indicator */
  .container .checkmark:after {
    left: 9px;
    top: 5px;
    width: 5px;
    height: 10px;
    border: solid white;
    border-width: 0 3px 3px 0;
    -webkit-transform: rotate(45deg);
    -ms-transform: rotate(45deg);
    transform: rotate(45deg);
  }
</style>
