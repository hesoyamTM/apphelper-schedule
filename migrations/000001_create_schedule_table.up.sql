CREATE TABLE IF NOT EXISTS public.schedules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(50) NOT NULL,
    group_id uuid NOT NULL,
    student_id uuid NOT NULL,
    trainer_id uuid NOT NULL,
    start_date timestamp NOT NULL,
    end_date timestamp NOT NULL,
);

