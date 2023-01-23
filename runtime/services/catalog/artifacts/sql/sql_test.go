package sql

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sanitizeQuery(t *testing.T) {
	sanitizeTests := []struct {
		title  string
		input  string
		output string
	}{
		{
			"removes comments, unused whitespace, and ;",
			`
			-- whatever this is
			SELECT * from         whatever;
			-- another extraneous comment.
			`,
			"SELECT * from whatever",
		},
		{
			"option to not lowercase a query",
			`
			-- whatever this is
			SELECT * from         whateveR;
			-- another extraneous comment.
			`,
			"SELECT * from whateveR",
		},
		{
			"removes extraneous spaces from columns",
			`
			-- whatever this is
			SELECT 1, 2,     3 from         whateveR;
			-- another extraneous comment.        
        	`,
			"SELECT 1,2,3 from whateveR",
		},
		{
			"multi line comments",
			`
			-- whatever this is
			-- second
			SELECT 1, 2,     3 from         whateveR;
			-- another extraneous comment.        
        	`,
			"SELECT 1,2,3 from whateveR",
		},
		{
			"materialize comment",
			`
			-- @materialize: true
			SELECT 1, 2,     3 from         whateveR;
			-- another extraneous comment.        
        	`,
			"SELECT 1,2,3 from whateveR",
		},
		{
			"materialize comment",
			`
			-- @materialize:  
			-- true
			SELECT 1, 2,     3 from         whateveR;
			-- another extraneous comment.        
        	`,
			"SELECT 1,2,3 from whateveR",
		},
		{
			"lines without comment will be kept", // will fail the model validation later
			`
			-- @materialize:  
			  true
			SELECT 1, 2,     3 from         whateveR;
			-- another extraneous comment.        
        	`,
			"true SELECT 1,2,3 from whateveR",
		},
	}

	for _, sanitizeTest := range sanitizeTests {
		t.Run(sanitizeTest.title, func(t *testing.T) {
			require.Equal(t, sanitizeTest.output, sanitizeQuery(sanitizeTest.input))
		})
	}
}

func Test_parseMaterializationInfo(t *testing.T) {
	sanitizeTests := []struct {
		title  string
		input  string
		output runtimev1.Model_Materialize
	}{
		{
			"materialize true",
			`
			-- @materialize: true 
			SELECT * from         whatever;
			-- another extraneous comment.
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"materialize inferred",
			`
			-- @materialize: inferred 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INFERRED,
		},
		{
			"materialize false",
			`
			-- @materialize: false 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_FALSE,
		},
		{
			"materialize invalid value",
			`
			-- @materialize: random 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INVALID,
		},
		{
			"parse invalid value",
			`
			-- @materialize: tru 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INVALID,
		},
		{
			"parse invalid value",
			`
			-- @materialize:  
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INVALID,
		},
		{
			"parse mix cap values",
			`
			-- @materialize: TruE 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"parse surrounding comments",
			`
			-- some comment.
			-- @materialize: inferred -- another comment
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INFERRED,
		},
		{
			"parse single space before colon",
			`
			-- @materialize : true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"parse single tab before colon",
			`
			-- @materialize	: true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"parse single tab after and before colon",
			`
			-- @materialize	:	true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"parse multiple tab after colon",
			`
			-- @materialize:		true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"parse mix of tabs and space after colon",
			`
			-- @materialize:		 true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"parse extra spaces after colon",
			`
			-- @materialize	:  true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"fail parsing extra spaces before colon",
			`
			-- @materialize  : true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_UNSPECIFIED,
		},
		{
			"fail parsing mix of space and tab before colon",
			`
			-- @materialize	 : true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_UNSPECIFIED,
		},
		{
			"fail parsing materialize caps keyword",
			`
			-- @Materialize: true 
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_UNSPECIFIED,
		},
		{
			"parse materialize caps value",
			`
			-- @materialize: True
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_TRUE,
		},
		{
			"fail parsing new line value",
			`
			-- @materialize: 
			true
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INVALID,
		},
		{
			"fail parsing new line value with comment",
			`
			-- @materialize: 
			-- true
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_INVALID,
		},
		{
			"parse incomplete comment",
			`
			-- @material
			SELECT * from whatever;
			`,
			runtimev1.Model_MATERIALIZE_UNSPECIFIED,
		},
		{
			"parse materialize comment not present",
			"SELECT * from whatever;",
			runtimev1.Model_MATERIALIZE_UNSPECIFIED,
		},
	}

	for _, sanitizeTest := range sanitizeTests {
		t.Run(sanitizeTest.title, func(t *testing.T) {
			require.Equal(t, sanitizeTest.output, parseMaterializationInfo(sanitizeTest.input))
		})
	}
}
