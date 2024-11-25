<script lang="ts">
  import { EnvironmentType } from "./types";

  export let environment: string;
  export let name: string;

  // NOTE: if environment is empty, the variable is shared for all environments
  function getEnvironmentLabel(environment: string) {
    if (environment === EnvironmentType.UNDEFINED) {
      return "Development, Production";
    }
    if (environment === EnvironmentType.DEVELOPMENT) {
      return "Development";
    }
    if (environment === EnvironmentType.PRODUCTION) {
      return "Production";
    }
    return "";
  }

  $: environmentLabel = getEnvironmentLabel(environment);
</script>

<div class="flex flex-col">
  <code class="text-sm text-gray-800 font-medium">{name}</code>
  <span class="text-xs text-gray-500 font-normal">{environmentLabel}</span>
</div>
