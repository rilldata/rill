package sql

import (
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
		output MaterializationInfo
	}{
		{
			"materialize true",
			`
			-- @materialize: true 
			SELECT * from         whatever;
			-- another extraneous comment.
			`,
			MaterializeTrue,
		},
		{
			"materialize inferred",
			`
			-- @materialize: inferred 
			SELECT * from whatever;
			`,
			MaterializeInferred,
		},
		{
			"materialize false",
			`
			-- @materialize: false 
			SELECT * from whatever;
			`,
			MaterializeFalse,
		},
		{
			"materialize invalid value",
			`
			-- @materialize: random 
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"parse invalid value",
			`
			-- @materialize: tru 
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"parse invalid value",
			`
			-- @materialize:  
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"parse spaces before",
			`
			  	-- @materialize: true  
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse multiple tags, use first",
			`
			  	-- @materialize: true -- @materialize: false
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse multiple tags, use first",
			`
			  	-- @materialize: t -- @materialize: false
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"parse multiple tags, use first",
			`
			-- @materialize: 
			-- @materialize: false
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"parse multiple tags, use first",
			`
			-- @materialize
			-- @materialize: inferred
			SELECT * from whatever;
			`,
			MaterializeInferred,
		},
		{
			"parse mix cap values",
			`
			-- @materialize: TruE 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse surrounding comments",
			`
			-- some comment.
			-- @materialize: inferred -- another comment
			SELECT * from whatever;
			`,
			MaterializeInferred,
		},
		{
			"parse single space before colon",
			`
			-- @materialize : true 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse single tab before colon",
			`
			-- @materialize	: true 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse single tab after and before colon",
			`
			-- @materialize	:	true 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse multiple tab after colon",
			`
			-- @materialize:		true 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse mix of tabs and space after colon",
			`
			-- @materialize:		 true 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"parse extra spaces after colon",
			`
			-- @materialize	:  true 
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"fail parsing extra spaces before colon",
			`
			-- @materialize  : true 
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"fail parsing tag on new line",
			`
			-- 
			@materialize: true 
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"fail parsing value on new line as comment",
			`
			-- @materialize 
			-- :true
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"fail parsing value on new line",
			`
			-- @materialize
			:true
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"fail parsing mix of space and tab before colon",
			`
			-- @materialize	 : true 
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"fail parsing materialize caps keyword",
			`
			-- @Materialize: true 
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"parse materialize caps value",
			`
			-- @materialize: True
			SELECT * from whatever;
			`,
			MaterializeTrue,
		},
		{
			"fail parsing new line value",
			`
			-- @materialize: 
			true
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"fail parsing new line value with comment",
			`
			-- @materialize: 
			-- true
			SELECT * from whatever;
			`,
			MaterializeInvalid,
		},
		{
			"parse incomplete comment",
			`
			-- @material
			SELECT * from whatever;
			`,
			MaterializeUnspecified,
		},
		{
			"parse materialize comment not present",
			"SELECT * from whatever;",
			MaterializeUnspecified,
		},
	}

	for _, sanitizeTest := range sanitizeTests {
		t.Run(sanitizeTest.title, func(t *testing.T) {
			require.Equal(t, sanitizeTest.output, parseMaterializationInfo(sanitizeTest.input))
		})
	}
}
