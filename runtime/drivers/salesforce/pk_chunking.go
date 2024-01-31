package salesforce

import (
	"regexp"
	"sort"
	"strings"
)

func isPKChunkingEnabled(bulkJob *bulkJob) bool {
	return bulkJob.pkChunkSize > 0 && isPKChunkingEnabledObject(bulkJob.objectName)
}

// pk chunking only works for certain standard objects, custom objects and share/history of those
func isPKChunkingEnabledObject(objectName string) bool {
	standardObjectPKChunkingEnabled := []string{"account", "accounthistory", "accountshare", "campaign", "campaignhistory", "campaignmember", "campaignmemberhistory", "campaignmembershare", "campaignshare", "case", "casehistory", "caseshare", "contact", "contacthistory", "contactshare", "event", "eventhistory", "eventrelation", "eventrelationhistory", "eventrelationshare", "eventshare", "lead", "leadhistory", "leadshare", "opportunity", "opportunityhistory", "opportunityshare", "task", "taskhistory", "taskshare", "user", "userhistory", "usershare"}

	isCustomObject, err := regexp.MatchString("__c$", objectName)
	if err != nil {
		panic("Regex errored out with " + err.Error())
	}
	isShareHistoryCustomObject, err := regexp.MatchString("(__Share|__History)$", objectName)
	if err != nil {
		panic("Regex errored out with " + err.Error())
	}
	isHistoricalTrendingObject, err := regexp.MatchString("_hd$", objectName)
	if err != nil {
		panic("Regex errored out with " + err.Error())
	}

	return contains(standardObjectPKChunkingEnabled, objectName) || isCustomObject || isShareHistoryCustomObject || isHistoricalTrendingObject
}

// performs a binary search for a given string
func contains(values []string, val string) bool {
	val = strings.ToLower(val)
	index := sort.SearchStrings(values, val)

	return index < len(values) && values[index] == val
}

// if a object has Share or History (__Share, __History for custom) suffix, it likely has a parent object, which should be queried when using pk chunking
func parentObject(objectName string) string {
	var parent string
	regex := regexp.MustCompile("(__Share|Share|__History|History)$")
	indexes := regex.FindStringIndex(objectName)

	if indexes != nil {
		start, end := indexes[0], indexes[1]
		suffix := objectName[start:end]
		isCustomObject := suffix[0:2] == "__"

		parent = objectName[:start]
		if isCustomObject {
			parent += "__c"
		}
	}

	return parent
}
