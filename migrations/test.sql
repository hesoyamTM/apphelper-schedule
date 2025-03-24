SELECT groups.name, schedules.group_id, schedules.student_id, schedules.trainer_id, schedules.date
		FROM schedules
		INNER JOIN groups ON groups.id = schedules.group_id
		WHERE schedules.trainer_id = 49