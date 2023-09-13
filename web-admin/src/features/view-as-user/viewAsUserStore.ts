import { writable } from "svelte/store";
import type { V1User } from "../../client";

export const viewAsUserStore = writable<V1User | null>(null);
