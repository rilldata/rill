import "../src/app.css";
import "../static/fonts/fonts.css";
import { fn } from "@storybook/test";
import type { Preview } from "@storybook/svelte";

const preview: Preview = {
  parameters: {
    actions: { "on:click": fn() },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/,
      },
    },
  },
};

export default preview;
