select
    "Date" as date,
    "PROPLetter" as letter,
    "Kind" as kind,
    "HowPlaced" as proposed_by,
    "PROPTitle" as title,
    "Description" as description,
    "Voter_Information_Pamphlet" as pamphlet_url,
    "PASS_FAIL" = 'P' as passed,
    "Vote_Counts_Yes" as votes_yes,
    "Vote_Counts_No" as votes_no,
    "Percent_Vote_Yes" as votes_yes_pct,
    "Percent_Vote_No" as votes_no_pct,
    "Percent_Required_To_Pass" as pass_pct
from sf_props_source
