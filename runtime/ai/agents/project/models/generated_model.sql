SELECT 
    generate_series AS event_id,
    current_date - (random() * 120)::int AS event_date,
    ['Meeting', 'Workshop', 'Webinar', 'Conference'][((random() * 4)::int + 1)] AS event_type,
    'USER_' || (generate_series % 500)::text AS organizer_id,
    (random() * 50 + 1)::int AS participants_count,
    (random() * 8 + 1)::int AS duration_hours,
    ((random() * 8 + 1)::int) * 60 AS duration_minutes, -- New column for duration in minutes
    (random() < 0.5) AS is_virtual,
    (random() * 100)::decimal(10,2) AS cost,
    ['Room A', 'Room B', 'Room C', 'Room D'][((random() * 4)::int + 1)] AS location,
    (random() * 5)::int AS priority_level,
    ['Draft', 'Scheduled', 'Completed', 'Cancelled'][((random() * 4)::int + 1)] AS status,
    (random() * 1000)::decimal(10,2) AS budget,
    (random() * 10 + 1)::int AS agenda_items,
    ['Internal', 'External'][((random() * 2)::int + 1)] AS audience_type,
    (random() < 0.3) AS is_recurring,
    'SUBJ_' || (generate_series % 100)::text AS subject_id,
    ['Confirmed', 'Tentative', 'Declined'][((random() * 3)::int + 1)] AS response_status,
    (random() * 100)::decimal(10,2) AS satisfaction_score,
    (random() * 20)::int AS follow_up_tasks,
    (random() * 500)::decimal(10,2) AS revenue_generated,
    -- Added new column for total duration in minutes
    ((random() * 8 + 1)::int * 60) + (random() * 60)::int AS total_duration_minutes
FROM generate_series(1, 7000)