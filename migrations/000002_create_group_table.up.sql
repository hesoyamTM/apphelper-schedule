CREATE TABLE IF NOT EXISTS groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    trainer_id uuid NOT NULL,
    student_ids uuid[] NOT NULL,
    invitation_link VARCHAR(20) NOT NULL
);