import { createReadableFactoryWithSelector } from "$lib/redux-store/svelte-readables-wrapper";
import { store } from "$lib/redux-store/store-root";
import { selectBigNumberById } from "$lib/redux-store/big-number/big-number-selectors";

export const getBigNumberById = createReadableFactoryWithSelector(
  store,
  selectBigNumberById
);
