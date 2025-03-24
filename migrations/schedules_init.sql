CREATE TABLE IF NOT EXISTS schedules (
    id SERIAL,
    group_id INT NOT NULL,
    student_id INT NOT NULL,
    trainer_id INT NOT NULL,
    date DATE NOT NULL
)