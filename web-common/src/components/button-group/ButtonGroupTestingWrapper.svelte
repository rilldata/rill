<script lang="ts">
  /**
   * Adding this component for testing purposes b/c it was not clear
   * how to render a component with children using @testing-library.
   * Doing it this way was the most expedient way to move forward.
   */

  import { ButtonGroup, SubButton } from "./index";

  export let values: number[] = [];
  export let selected: number[] = [];
  export let disabled: number[] = [];

  const testTooltips = {
    selected: "selected tt",
    unselected: "unselected tt",
    disabled: "disabled tt",
  };
</script>

<ButtonGroup {...{ selected, disabled }} on:subbutton-click>
  {#each values as key}
    <SubButton value={key} tooltips={testTooltips}>
      button-{key}
    </SubButton>
  {/each}
</ButtonGroup>
