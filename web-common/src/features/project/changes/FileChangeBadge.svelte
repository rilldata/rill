<script lang="ts">
  import type { GitDiffResponse_GitFileStatus } from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { status }: { status: GitDiffResponse_GitFileStatus | undefined } =
    $props();

  // The v2 client serializes proto enums as their JSON string names (e.g. "GIT_FILE_STATUS_ADDED").
  const badges: Record<
    string,
    { letter: string; class: string; label: string }
  > = {
    GIT_FILE_STATUS_ADDED: {
      letter: "A",
      class: "badge-added",
      label: m.edit_changes_status_added(),
    },
    GIT_FILE_STATUS_MODIFIED: {
      letter: "M",
      class: "badge-modified",
      label: m.edit_changes_status_modified(),
    },
    GIT_FILE_STATUS_DELETED: {
      letter: "D",
      class: "badge-deleted",
      label: m.edit_changes_status_deleted(),
    },
    GIT_FILE_STATUS_RENAMED: {
      letter: "R",
      class: "badge-renamed",
      label: m.edit_changes_status_renamed(),
    },
  };

  const badge = $derived(
    badges[status as unknown as string] ?? badges.GIT_FILE_STATUS_MODIFIED,
  );
</script>

<span class="badge {badge.class}" title={badge.label}>{badge.letter}</span>

<style lang="postcss">
  .badge {
    @apply flex-none text-[0.625rem] leading-none px-1 py-0.5 rounded;
    @apply font-mono font-medium;
  }

  .badge-added {
    @apply bg-primary-100 text-primary-800;
  }

  .badge-modified {
    @apply bg-yellow-100 text-yellow-700;
  }

  .badge-deleted {
    @apply bg-red-100 text-red-700;
  }

  .badge-renamed {
    @apply bg-secondary-100 text-secondary-800;
  }
</style>
