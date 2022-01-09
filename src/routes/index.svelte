<script context="module">
    // Redirect to /model for now
    export async function load() {
        return {
            status: 302,
            redirect: "/model"
        };
    }
</script>