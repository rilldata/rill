<script lang="ts">
  import "./wizard.css";

  export let managementType: "rill-managed" | "self-managed";
  export let onSelectManagementType: (
    managementType: "rill-managed" | "self-managed",
  ) => void;

  const managementTypeOptions = ["rill-managed", "self-managed"];
</script>

<div class="management-toggle">
  {#each managementTypeOptions as option (option)}
    <button
      class="option"
      class:selected={managementType === option}
      on:click={() => onSelectManagementType(option)}
    >
      {option.charAt(0).toUpperCase() + option.slice(1)}
    </button>
  {/each}
</div>

<style lang="postcss">
  .management-toggle {
    @apply flex flex-row gap-x-4;
    justify-content: center;
    margin-bottom: 1rem;
  }

  button {
    @apply w-[196px] h-[40px];
    @apply border border-slate-300 rounded;
  }
</style>
