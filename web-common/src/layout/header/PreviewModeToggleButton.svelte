<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Pencil, Play } from "lucide-svelte";

  export let mode: "Preview" | "Edit" = "Preview";
  export let href: string;
  export let disabled = false;
</script>

<Button type="secondary" preload={false} compact label={mode} {href} {disabled}>
  <div class="flex gap-x-1 items-center">
    {#if mode === "Preview"}
      <Play size={14} />
    {:else}
      <Pencil size={14} />
    {/if}
    {mode}
  </div>
</Button>
