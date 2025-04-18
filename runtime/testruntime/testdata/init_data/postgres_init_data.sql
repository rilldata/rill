CREATE TYPE country AS ENUM ('IND', 'AUS', 'SA', 'NZ');
CREATE TABLE all_datatypes (
	id serial PRIMARY KEY,
	uuid UUID,
	name text,
	age integer,
	is_married boolean,
	date_of_birth date,
	time_of_day time,
	created_at timestamp,
	personal_info json,
	personal_info2 jsonb,
	is_alive bit,
	binary_data bit varying,
	gender character,
	gender_full character varying,
	nickname bpchar(10),
	num_of_dependents smallint,
	biography text,
	last_login timestamptz,
	weight float4,
	height float8,
	sibling_rank int2,
	credit_score int4,
	net_worth int8,
	salary_history int8[],
	login_history timestamptz[],
	emp_salary NUMERIC,
	country country
);

INSERT INTO all_datatypes (uuid, name, age, is_married, date_of_birth, time_of_day, created_at, personal_info, personal_info2, is_alive, binary_data, gender, gender_full, nickname, num_of_dependents, biography, last_login, weight, height, sibling_rank, credit_score, net_worth, salary_history, login_history, emp_salary, country)
VALUES
	('8a25ac46-8ad6-4415-9a2e-12aa3962c144', 'John Doe', 30, true, '1983-03-08', '12:35:00', '2023-09-12 12:46:55', '{"hobbies": "Travel, Tech"}', '{"job": "Software Engineer"}', b'1', b'10101010', 'M', 'Male', 'abcd', 2, 'John is a software engineer who loves to travel and explore new places.', '2023-09-12 12:46:55+05:30', 75.4, 180.5, 1, 720, 1234567, Array[1234567, 7654312], Array[timestamp '2023-09-12 12:46:55+05:30', timestamp '2023-10-12 12:46:55+05:30'], 385000.71, 'IND'),
	('ec773cd0-8edc-419a-9d57-1815aaee2f01', 'Alice Smith', 25, false, '1998-07-15', '08:20:00', '2023-08-10 10:20:30', '{"hobbies": "Reading, AI"}', '{"job": "Data Analyst"}', b'0', b'11110000', 'F', 'Female', 'wxyz', 0, 'Alice is a data analyst with a passion for AI and machine learning.', '2023-08-10 10:20:30+05:30', 62.3, 167.2, 2, 680, 8765432, Array[8765432, 2345678], NULL, 550000.12, 'AUS'),
	('ddb115ff-8da4-4b36-b1b5-1f58123c1552', 'Bob Brown', 40, NULL, '1982-01-22', '14:45:00', '2023-07-15 15:45:20', '{"hobbies": "Cycling, Management"}', '{"job": "Project Manager"}', NULL, b'11001100', 'M', 'Male', 'mnop', 3, 'Bob is a project manager with 15 years of experience.', '2023-07-15 15:45:20+05:30', 85.2, 175.0, 3, 710, 6543210, NULL, Array[timestamp '2023-07-15 15:45:20+05:30'], NULL, 'NZ'),
	('5cf3d245-3d9b-4baf-b0f3-9c2f29150c57', 'Sophia Davis', 35, true, '1987-11-30', '09:30:00', '2023-06-20 20:10:05', '{"hobbies": "Design, Art"}', '{"job": "Designer"}', b'1', NULL, 'F', 'Female', 'qrst', 1, 'Sophia is a designer who enjoys creating user-friendly experiences.', '2023-06-20 20:10:05+05:30', 58.9, 160.0, 4, 750, 9876543, Array[9999, 8888], NULL, 6500000.65, 'SA'),
	('c13da985-454a-48f1-9c35-e4281f918a77', 'Emma White', 28, false, '1995-02-14', '10:05:00', '2023-05-25 09:55:10', '{"hobbies": "Research, Science"}', '{"job": "Researcher"}', b'0', b'10101010', 'F', 'Female', 'uvwx', 0, 'Emma is a researcher focused on environmental science.', '2023-05-25 09:55:10+05:30', 65.0, 170.8, 5, 690, 7890123, Array[7890123, 2109876], Array[timestamp '2023-05-25 09:55:10+05:30'], 4800000.98, 'IND');
