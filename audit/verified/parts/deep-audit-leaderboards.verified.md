# Verified audit: deep-audit-leaderboards

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31)

Exhaustive source finding count: **13 numbered findings** (`LEAD-1` through `LEAD-13`). Every heading appears exactly once.

## Pending

| ID | Original severity | Priority | Concise title | Current code evidence | Recommended fix |
|---|---|---|---|---|---|
| LEAD-1 | medium | P1 | Unsafe URI schemes pass through makeHref | `dashboards/dashboard-utils.ts:155-181` still returns any scheme-bearing value | Central protocol allowlist; reject unsafe schemes before rendering links. |
| LEAD-2 | medium | P1 | Exact zero is rendered as no data | `leaderboard/LeaderboardRow.svelte:261,305,335` still uses value truthiness | Replace with explicit null checks for measure/delta cells. |
| LEAD-3 | medium | P1 | Percent-of-total can be NaN/Infinity | `leaderboard/leaderboard-utils.ts:105-106,215-216` still divides by zero/undefined totals | Require a finite non-zero total. |
| LEAD-4 | medium | P1 | Selected values vanish when only some lack data | `leaderboard/Leaderboard.svelte:272-279` still uses an all-or-nothing placeholder fallback | Merge result rows with placeholders for every missing selected value. |
| LEAD-5 | medium | P1 | Type mismatch can duplicate selected row | `leaderboard/leaderboard-utils.ts:189-193` still deletes from Set by strict identity but highlights with coercing comparator | Prune with the same canonical comparator/type normalization. |
| LEAD-6 | medium | P1 | Dimension table selection lacks legacy coercion | `dimension-table/DimensionTable.svelte:66` still compares differently from leaderboard compatibility logic | Reuse one canonical dimension-value comparator. |
| LEAD-7 | medium | P2 | Global select-all hijacks text editing | `dimension-table/DimensionDisplay.svelte:233` still intercepts Ctrl/Cmd+A without editable-target checks | Ignore input/textarea/contenteditable and require table focus. |
| LEAD-8 | medium | P2 | Scroll range recreates table virtualizers | `dimension-table/DimensionTable.svelte:103` still rebuilds virtualizers from feedback state | Create once and update measurements/options imperatively. |
| LEAD-9 | medium | P2 | Overscan nearly disables row virtualization | `dimension-table/DimensionTable.svelte:54,113` still defaults to 120 for a 250-row table | Reduce/adapt overscan and measure scroll performance. |
| LEAD-10 | medium | P2 | Selection stores are recreated on broad state change | `leaderboard/LeaderboardDisplay.svelte:104` still constructs inline stores from broad dashboard state | Memoize/select only relevant dimension filter state. |
| LEAD-11 | medium | P2 | Column widths remain based on initial empty rows | `dimension-table/DimensionTable.svelte:78` still estimates once before data and never recomputes | Recompute after first data and on material schema/content changes. |
| LEAD-12 | low | P2 | Shared table config leaks header height | `dimension-table/DimensionTable.svelte:85` still mutates singleton `DIMENSION_TABLE_CONFIG` | Clone config per instance or reset all mutations. |
| LEAD-13 | low | P2 | Leaderboard observer is never disconnected | `leaderboard/Leaderboard.svelte:96-110` still lacks observer cleanup | Disconnect in `onDestroy`. |

## Fixed / remove

None.

## Partial / needs test

None.

Counts: **13 pending (6 P1, 7 P2), 0 fixed, 0 partial = 13**.

## Machine-readable inventory

```text
PENDING|deep-audit-leaderboards.md|LEAD-1|medium|P1|-|Unsafe URI schemes|dashboard-utils.ts:155-181
PENDING|deep-audit-leaderboards.md|LEAD-2|medium|P1|-|Zero shown as no-data|LeaderboardRow.svelte:261,305,335
PENDING|deep-audit-leaderboards.md|LEAD-3|medium|P1|-|NaN/Infinity percent total|leaderboard-utils.ts:105-106,215-216
PENDING|deep-audit-leaderboards.md|LEAD-4|medium|P1|-|Selected no-data values vanish|Leaderboard.svelte:272-279
PENDING|deep-audit-leaderboards.md|LEAD-5|medium|P1|-|Selected row duplicates by type|leaderboard-utils.ts:189-193
PENDING|deep-audit-leaderboards.md|LEAD-6|medium|P1|-|Selection compatibility mismatch|DimensionTable.svelte:66
PENDING|deep-audit-leaderboards.md|LEAD-7|medium|P2|-|Global select-all hijack|DimensionDisplay.svelte:233
PENDING|deep-audit-leaderboards.md|LEAD-8|medium|P2|-|Virtualizers recreated on scroll|DimensionTable.svelte:103
PENDING|deep-audit-leaderboards.md|LEAD-9|medium|P2|-|Excessive overscan|DimensionTable.svelte:54,113
PENDING|deep-audit-leaderboards.md|LEAD-10|medium|P2|-|Selection stores recreated|LeaderboardDisplay.svelte:104
PENDING|deep-audit-leaderboards.md|LEAD-11|medium|P2|-|Widths estimated before data|DimensionTable.svelte:78
PENDING|deep-audit-leaderboards.md|LEAD-12|low|P2|-|Singleton config leaks height|DimensionTable.svelte:85
PENDING|deep-audit-leaderboards.md|LEAD-13|low|P2|-|Observer not disconnected|Leaderboard.svelte:96-110
```
