<!-- @component
This component will unmount the slotted text if the inspector width is less than
the threshold specified in collapseInspectorCTAButton, causing a cute
horizontal slide animation
-->
<script lang="ts">
  import { slideRight } from "@rilldata/web-common/lib/transitions";
  import { collapseInspectorCTAButton } from "../../application-config";

  export let width: number;
</script>

{#if collapseInspectorCTAButton(width)}<div
    transition:slideRight|local={{
      duration: 400,
      rightOffset: 8,
    }}
  >
    <slot />
  </div>
{/if}
