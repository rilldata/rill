import { mount } from "svelte";
import App from "./App.svelte";

const target = document.getElementById("app");
if (!target) throw new Error("rill chart widget: #app container missing");

mount(App, { target });
