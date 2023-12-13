declare namespace svelteHTML {
  interface HTMLAttributes {
    // Used for copy action `shift-click-actions.ts`
    "on:shift-click"?: (event: CustomEvent) => void;
    "on:command-click"?: (event: CustomEvent) => void;
  }

  interface SVGAttributes {
    "on:scrub-start"?: (event: CustomEvent) => void;
    "on:scrub-end"?: (event: CustomEvent) => void;
    "on:scrub-move"?: (event: CustomEvent) => void;
  }
}
