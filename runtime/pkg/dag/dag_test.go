package dag

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDAG_Add(t *testing.T) {
	d := NewDAG()

	d.Add("A0", []string{})
	d.Add("B1", []string{"B0", "C0"})
	d.Add("B2", []string{"A1", "B1"})
	// A0  B0  C0
	//     |  /
	// A1  B1
	//   \ |
	//     B2
	require.Equal(t, []string{}, d.GetChildren("A0"))
	require.Equal(t, []string{"B1", "B2"}, d.GetChildren("B0"))
	require.Equal(t, []string{"B1", "B2"}, d.GetChildren("C0"))
	require.Equal(t, []string{"B2"}, d.GetChildren("A1"))
	require.Equal(t, []string{"B2"}, d.GetChildren("B1"))

	d.Add("A1", []string{"A0", "B0"})
	d.Add("A2", []string{"C0"})
	// A0  B0  C0
	// | / | / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2
	require.Equal(t, []string{"A1", "B2"}, d.GetChildren("A0"))
	require.ElementsMatch(t, []string{"A1", "B1", "B2"}, d.GetChildren("B0"))
	require.ElementsMatch(t, []string{"B1", "A2", "B2"}, d.GetChildren("C0"))
	require.Equal(t, []string{"B2"}, d.GetChildren("A1"))
	require.Equal(t, []string{"B2"}, d.GetChildren("A1"))

	d.Add("A1", []string{"C0"})
	d.Add("B1", []string{"C0"})
	// A0   C0   B0
	//    / / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2
	require.Equal(t, []string{}, d.GetChildren("A0"))
	require.Equal(t, []string{}, d.GetChildren("B0"))
	require.ElementsMatch(t, []string{"B1", "A2", "A1", "B2"}, d.GetChildren("C0"))
}

func TestDAG_DeleteButBranchRetained(t *testing.T) {
	d := getTestDAG()
	d.Delete("A0")
	require.Equal(t, []string{"A1", "B2"}, d.GetChildren("A0"))

	d.Delete("A1")
	require.Equal(t, []string{"A1", "B2"}, d.GetChildren("A0"))

	d.Add("A1", []string{"A0"})
	d.Delete("B2")
	require.Equal(t, []string{"A1"}, d.GetChildren("A0"))
}

func TestDAG_DeleteBranch(t *testing.T) {
	d := getTestDAG()
	d.Delete("A0")
	d.Delete("A1")
	require.Equal(t, []string{"A1", "B2"}, d.GetChildren("A0"))

	d.Delete("B2")
	require.Equal(t, []string{}, d.GetChildren("A0"))
}

func getTestDAG() *DAG {
	d := NewDAG()
	d.Add("A0", []string{})
	d.Add("B1", []string{"B0", "C0"})
	d.Add("B2", []string{"A1", "B1"})
	d.Add("A1", []string{"A0", "B0"})
	d.Add("A2", []string{"C0"})
	// A0  B0  C0
	// | / | / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2
	return d
}

func TestCyclicDAG(t *testing.T) {
	d := NewDAG()
	d.Add("A0", []string{"B1"})
	n, err := d.Add("B1", []string{"A0"})
	require.Error(t, err)

	d.Delete("A0")
	d.Delete("B1")
	d.Add("A0", []string{})
	d.Add("B1", []string{"A0"})
	d.Add("B2", []string{"B1"})
	n, err = d.Add("B0", []string{"B1"})
	// A0
	// |
	// B1
	// |  \
	// B2 B0

	require.NoError(t, err)
	require.Equal(t, "B0", n.Name)
	require.Equal(t, []string{"B1", "B2", "B0"}, d.GetChildren("A0"))

	// A0 ----
	// |      |
	// B1     |
	// |  \   |
	// B2 B0 -
	_, err = d.Add("A0", []string{"B0"})
	fmt.Println(err)
	require.Error(t, err)
}
