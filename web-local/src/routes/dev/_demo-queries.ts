export const queries = [
  `
WITH x AS (select a, b, c, d, whatevewr from table)
   select     a, b+c as   next_val, whatever        
   
   
   from x

`,
  `
	WITH x as (select * from x0),
	y as (select count(*) as count, category from x0 INNER JOIN y0 ON y0.id = x0.y_id GROUP BY category)
	SELECT 
		index, 
		length(bits) AS bitlength,
		created_date,
		user_agent,
		category,
		y.count AS count
	FROM x
		INNER JOIN y ON y.category = x.category
`,

  `WITH dataset AS (
	SELECT epoch(revenue_usd) as revenue_usd FROM revenue_transactions_cleaned
), S AS (
	SELECT 
		min(revenue_usd) as minVal,
		max(revenue_usd) as maxVal,
		(max(revenue_usd) - min(revenue_usd)) as range
		FROM dataset
), values AS (
	SELECT revenue_usd as value from dataset
	WHERE revenue_usd IS NOT NULL
), buckets AS (
	SELECT
		range as bucket,
		(range - 1) * (select range FROM S) / 40 + (select minVal from S) as low,
		(range) * (select range FROM S) / 40 + (select minVal from S) as high
	FROM range(0, 40, 1)
)
, histogram_stage AS (
	SELECT
		bucket,
		low,
		high,
		count(values.value) as count
	FROM buckets
	LEFT JOIN values ON (values.value BETWEEN low and high)
	GROUP BY bucket, low, high
	ORDER BY BUCKET
)
SELECT 
	bucket,
	low,
	high,
	CASE WHEN high = (SELECT max(high) from histogram_stage) THEN count + 1 ELSE count END AS count
	FROM histogram_stage;`,
  // gitlab dbt
  `
WITH detection_rule AS (
 
    SELECT *
    FROM data_detection_rule
 
), rule_run_detail AS (
  
   SELECT 
        rule_id,
        processed_record_count,
        passed_record_count,
        failed_record_count,
        ((passed_record_count/processed_record_count)*100) AS percent_of_records_passed,
        ((failed_record_count/processed_record_count)*100) AS percent_of_records_failed,
        rule_run_date,
        type_of_data
   FROM product_data_detection_run_detail
 
), final AS (
 
    SELECT DISTINCT
        detection_rule.rule_id,
        detection_rule.rule_name,
        detection_rule.rule_description,
        rule_run_detail.rule_run_date,
        rule_run_detail.percent_of_records_passed,
        rule_run_detail.percent_of_records_failed,
        IFF(percent_of_records_passed > threshold, TRUE, FALSE) AS is_pass,
        rule_run_detail.type_of_data
    FROM rule_run_detail
    LEFT OUTER JOIN  detection_rule ON
    rule_run_detail.rule_id = detection_rule.rule_id
 
)
SELECt * from final;
`,
];
