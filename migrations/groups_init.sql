CREATE TABLE IF NOT EXISTS groups (
    id SERIAL,
    name VARCHAR(50) NOT NULL,
    trainer_id INT NOT NULL,
    student_ids INT[] NOT NULL,
    invitation_link VARCHAR(20) NOT NULL
)