WITH
    SignUpFunnel AS (
        SELECT
            CAST( strptime(occurred_at, '%c') AS DATE) AS event_date,
            CASE WHEN location IN ('United States', 'United Kingdom') THEN location ELSE 'Other' END AS location_group,
            CAST(COUNT (DISTINCT user_id) AS FLOAT) AS total_signups,
            CAST(COUNT (DISTINCT CASE WHEN event_name = 'create_user' THEN user_id ELSE NULL END) AS FLOAT) AS total_create_user, 
            CAST(COUNT (DISTINCT CASE WHEN event_name = 'enter_email' THEN user_id ELSE NULL END) AS FLOAT) AS total_enter_email, 
            CAST(COUNT (DISTINCT CASE WHEN event_name = 'enter_info' THEN user_id ELSE NULL END) AS FLOAT) AS total_enter_info, 
            CAST(COUNT (DISTINCT CASE WHEN event_name = 'complete_signup' THEN user_id ELSE NULL END) AS FLOAT) AS total_complete_signup
        FROM
            yammerevents
        WHERE event_type = 'signup_flow'
        GROUP BY 1,2
    )
    SELECT
        *,
        total_create_user/total_signups AS create_user_rate,
        total_enter_email/total_signups AS enter_email_rate,
        total_enter_info/total_signups AS total_enter_info_rate,
        total_complete_signup/total_signups AS completion_rate
    FROM SignUpFunnel