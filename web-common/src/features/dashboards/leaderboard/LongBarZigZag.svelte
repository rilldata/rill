<script lang="ts">
  import { fly } from "svelte/transition";

  let zigZag =
    "M" +
    Array.from({ length: 7 })
      .map((_, i) => {
        return `${15 - 4 * (i % 2)} ${1.7 * (i * 2)}`;
      })
      .join(" L");
</script>

<div style="position: relative" transition:fly={{ duration: 200, x: 20 }}>
  <svg
    style="
    position:absolute;
    right: 0px;
    transform: translateY(-22px);
  "
    width="15"
    height="22"
  >
    <path d={zigZag} fill="white" />
  </svg>
</div>
