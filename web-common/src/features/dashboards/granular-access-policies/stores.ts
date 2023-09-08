import { writable } from "svelte/store";
import type { MockUser } from "./useMockUsers";

export const selectedMockUserStore = writable<MockUser | null>(null);
export const selectedMockUserJWT = writable<string | null>(null);
