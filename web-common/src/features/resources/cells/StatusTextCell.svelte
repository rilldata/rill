<script lang="ts">
  type Tone = "default" | "muted" | "success" | "destructive";

  export let label: string;
  export let tone: Tone = "default";
</script>

<span
  class="truncate"
  class:text-fg-tertiary={tone === "muted"}
  class:text-green-600={tone === "success"}
  class:text-destructive={tone === "destructive"}
>
  {label}
</span>
