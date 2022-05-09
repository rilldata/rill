import adapter from '@sveltejs/adapter-static';
import preprocess from 'svelte-preprocess';
import { resolve } from "path";
import typescript from '@rollup/plugin-typescript';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://github.com/sveltejs/svelte-preprocess
	// for more information about preprocessors
	preprocess: preprocess(),

	plugins: [
		typescript({ sourceMap: true }),
	],

	kit: {
		adapter: adapter(),

		vite: {
			resolve: {
				alias: {
					$common: resolve('./src/common'),
					$lib: resolve('./src/lib'),
				}
			}
		}
	}
};

export default config;
