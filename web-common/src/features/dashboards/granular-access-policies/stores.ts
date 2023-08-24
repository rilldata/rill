import { writable } from "svelte/store";
import type { MockUser } from "./useMockUsers";

export const selectedMockUserStore = writable<MockUser | null>(null);

export const mockUserHasNoAccessStore = writable(false);
