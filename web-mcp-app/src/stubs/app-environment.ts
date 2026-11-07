// Stand-in for SvelteKit's `$app/environment`: the widget is a plain Vite app, but
// some shared web-common modules (local storage stores) import this module.
export const browser = true;
export const building = false;
export const dev = import.meta.env.DEV;
export const version = "";
