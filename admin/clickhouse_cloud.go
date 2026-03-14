package admin

// CHCTiers maps ClickHouse Cloud cluster memory (GB per replica) to minimum Rill slots.
// Each tier's memory = slots * 2 GB.
var CHCTiers = []struct {
	MemoryGB int
	Slots    int
}{
	{8, 4},
	{12, 6},
	{16, 8},
	{32, 16},
	{64, 32},
	{120, 60},
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
