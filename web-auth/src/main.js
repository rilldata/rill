import Auth from "./components/Auth.svelte";

const app = new Auth({
	target: document.body,
	props: {
		// This gets populated by Auth0 runtime
		configParams: "@@config@@",
		// This get populated by RollUp
		cloudClientIDs: "%%cloudClientIDs%%"
	}
});

export default app;