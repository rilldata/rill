declare namespace svelteHTML {
  interface HTMLAttributes {
    // Used for copy action `shift-click-actions.ts`
    "on:shift-click"?: (event: CustomEvent) => void;
    "on:command-click"?: (event: CustomEvent) => void;
  }
}
