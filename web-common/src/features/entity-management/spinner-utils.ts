import { writable } from 'svelte/store';

export function isLoadingWithTimeout(
    isLoading: boolean = false,
    delay: number = 300,
) {
    const { subscribe, set } = writable(isLoading);

    let timeoutId: ReturnType<typeof setTimeout>;

    function setLoading(loadingState: boolean) {
        clearTimeout(timeoutId);
        if (loadingState) {
            timeoutId = setTimeout(() => set(true), delay);
        } else {
            set(false);
        }
    }

    return {
        subscribe,
        setLoading,
    };
}