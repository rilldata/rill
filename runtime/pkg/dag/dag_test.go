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
	childrensA0, err := d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{}, childrensA0)

	childrensB0, err := d.GetChildren("B0")
	require.NoError(t, err)
	require.Equal(t, []string{"B1", "B2"}, childrensB0)

	childrensC0, err := d.GetChildren("C0")
	require.NoError(t, err)
	require.Equal(t, []string{"B1", "B2"}, childrensC0)

	childrensA1, err := d.GetChildren("A1")
	require.NoError(t, err)
	require.Equal(t, []string{"B2"}, childrensA1)

	childrensB1, err := d.GetChildren("B1")
	require.NoError(t, err)
	require.Equal(t, []string{"B2"}, childrensB1)

	d.Add("A1", []string{"A0", "B0"})
	d.Add("A2", []string{"C0"})
	// A0  B0  C0
	// | / | / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2

	childrensA0, err = d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{"A1", "B2"}, childrensA0)

	childrensB0, err = d.GetChildren("B0")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"A1", "B1", "B2"}, childrensB0)

	childrensC0, err = d.GetChildren("C0")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"B1", "A2", "B2"}, childrensC0)

	childrensA1, err = d.GetChildren("A1")
	require.NoError(t, err)
	require.Equal(t, []string{"B2"}, childrensA1)

	childrensB1, err = d.GetChildren("A1")
	require.NoError(t, err)
	require.Equal(t, []string{"B2"}, childrensB1)

	d.Add("A1", []string{"C0"})
	d.Add("B1", []string{"C0"})
	// A0   C0   B0
	//    / / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2

	childrensA0, err = d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{}, childrensA0)

	childrensB0, err = d.GetChildren("B0")
	require.NoError(t, err)
	require.Equal(t, []string{}, childrensB0)

	childrensC0, err = d.GetChildren("C0")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"B1", "A2", "A1", "B2"}, childrensC0)
}

func TestDAG_DeleteButBranchRetained(t *testing.T) {
	d := getTestDAG()
	d.Delete("A0")
	childrensA0, err := d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{"A1", "B2"}, childrensA0)

	d.Delete("A1")
	childrensA0, err = d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{"A1", "B2"}, childrensA0)

	d.Add("A1", []string{"A0"})
	d.Delete("B2")
	childrensA0, err = d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{"A1"}, childrensA0)
}

func TestDAG_DeleteBranch(t *testing.T) {
	d := getTestDAG()
	d.Delete("A0")
	d.Delete("A1")
	childrensA0, err := d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{"A1", "B2"}, childrensA0)

	d.Delete("B2")
	childrensA0, err = d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{}, childrensA0)
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
	d.Add("A0", []string{})
	d.Add("B1", []string{"A0"})
	d.Add("B2", []string{"B1"})
	d.Add("B0", []string{"B1"})
	// fmt.Println(err)
	// A0
	// |
	// B1
	// |  \
	// B2 B0

	nodes, err := d.GetChildren("A0")
	require.NoError(t, err)
	require.Equal(t, []string{"B1", "B2", "B0"}, nodes)
	fmt.Println(nodes)

	d.Add("A0", []string{"B0"})
	nodes, err = d.GetChildren("A0")
	require.Error(t, err)
}
