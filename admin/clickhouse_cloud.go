package admin

// CHCTiers maps ClickHouse Cloud cluster memory (GB per replica) to minimum Rill slots.
// Each slot provides 1 CPU / 4 GB memory, so slots = ceil(memoryGB / 4).
var CHCTiers = []struct {
	MemoryGB int
	Slots    int
}{
	{8, 2},
	{12, 3},
	{16, 4},
	{32, 8},
	{64, 16},
	{120, 30},
}

// CHCMinSlotsForMemory returns the minimum Rill slots for a given CHC cluster memory (GB per replica).
// It picks the smallest tier whose memory is >= the cluster memory.
func CHCMinSlotsForMemory(memoryGB float64) int {
	for _, t := range CHCTiers {
		if memoryGB <= float64(t.MemoryGB) {
			return t.Slots
		}
	}
	return CHCTiers[len(CHCTiers)-1].Slots
}
