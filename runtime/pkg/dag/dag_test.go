package dag

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDAG_Add(t *testing.T) {
	d := NewDAG()

	_, err := d.Add("A0", []string{})
	require.NoError(t, err)
	_, err = d.Add("B1", []string{"B0", "C0"})
	require.NoError(t, err)
	_, err = d.Add("B2", []string{"A1", "B1"})
	require.NoError(t, err)
	// A0  B0  C0
	//     |  /
	// A1  B1
	//   \ |
	//     B2
	require.Equal(t, []string{}, d.GetDeepChildren("A0"))
	require.Equal(t, []string{"B1", "B2"}, d.GetDeepChildren("B0"))
	require.Equal(t, []string{"B1", "B2"}, d.GetDeepChildren("C0"))
	require.Equal(t, []string{"B2"}, d.GetDeepChildren("A1"))
	require.Equal(t, []string{"B2"}, d.GetDeepChildren("B1"))

	_, err = d.Add("A1", []string{"A0", "B0"})
	require.NoError(t, err)
	_, err = d.Add("A2", []string{"C0"})
	require.NoError(t, err)
	// A0  B0  C0
	// | / | / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2
	require.Equal(t, []string{"A1", "B2"}, d.GetDeepChildren("A0"))
	require.ElementsMatch(t, []string{"A1", "B1", "B2"}, d.GetDeepChildren("B0"))
	require.ElementsMatch(t, []string{"B1", "A2", "B2"}, d.GetDeepChildren("C0"))
	require.Equal(t, []string{"B2"}, d.GetDeepChildren("A1"))
	require.Equal(t, []string{"B2"}, d.GetDeepChildren("B1"))

	_, err = d.Add("A1", []string{"C0"})
	require.NoError(t, err)
	_, err = d.Add("B1", []string{"C0"})
	require.NoError(t, err)
	// A0   C0   B0
	//    / / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2
	require.Equal(t, []string{}, d.GetDeepChildren("A0"))
	require.Equal(t, []string{}, d.GetDeepChildren("B0"))
	require.ElementsMatch(t, []string{"B1", "A2", "A1", "B2"}, d.GetDeepChildren("C0"))
}

func TestDAG_DeleteButBranchRetained(t *testing.T) {
	d, err := getTestDAG()
	require.NoError(t, err)
	d.Delete("A0")
	require.Equal(t, []string{"A1", "B2"}, d.GetDeepChildren("A0"))

	d.Delete("A1")
	require.Equal(t, []string{"A1", "B2"}, d.GetDeepChildren("A0"))

	_, err = d.Add("A1", []string{"A0"})
	require.NoError(t, err)
	d.Delete("B2")
	require.Equal(t, []string{"A1"}, d.GetDeepChildren("A0"))
}

func TestDAG_DeleteBranch(t *testing.T) {
	d, err := getTestDAG()
	require.NoError(t, err)

	d.Delete("A0")
	d.Delete("A1")
	require.Equal(t, []string{"A1", "B2"}, d.GetDeepChildren("A0"))

	d.Delete("B2")
	require.Equal(t, []string{}, d.GetDeepChildren("A0"))
}

func getTestDAG() (*DAG, error) {
	d := NewDAG()
	_, err := d.Add("A0", []string{})
	if err != nil {
		return nil, err
	}
	_, err = d.Add("B1", []string{"B0", "C0"})
	if err != nil {
		return nil, err
	}
	_, err = d.Add("B2", []string{"A1", "B1"})
	if err != nil {
		return nil, err
	}
	_, err = d.Add("A1", []string{"A0", "B0"})
	if err != nil {
		return nil, err
	}
	_, err = d.Add("A2", []string{"C0"})
	if err != nil {
		return nil, err
	}
	// A0  B0  C0
	// | / | / |
	// A1  B1  |
	//   \ |   |
	//     B2  A2
	return d, nil
}

func TestCyclicDAG(t *testing.T) {
	d := NewDAG()
	n, err := d.Add("A0", []string{"B1"})
	require.NoError(t, err)
	require.Equal(t, "A0", n.Name)

	p, ok := n.Parents["B1"]
	require.Equal(t, true, ok)
	require.Equal(t, "B1", p.Name)

	n, err = d.Add("B1", []string{"A0"})
	require.Nil(t, n)
	require.Error(t, err)

	d.Delete("A0")
	d.Delete("B1")
	_, err = d.Add("A0", []string{})
	require.NoError(t, err)
	_, err = d.Add("B1", []string{"A0"})
	require.NoError(t, err)
	_, err = d.Add("B2", []string{"B1"})
	require.NoError(t, err)
	n, err = d.Add("B0", []string{"B1"})
	// A0
	// |
	// B1
	// |  \
	// B2 B0

	require.NoError(t, err)
	require.Equal(t, "B0", n.Name)
	require.ElementsMatch(t, []string{"B1", "B2", "B0"}, d.GetDeepChildren("A0"))

	// A0 ----
	// |      |
	// B1     |
	// |  \   |
	// B2 B0 -
	_, err = d.Add("A0", []string{"B0"})
	require.Error(t, err)
}
