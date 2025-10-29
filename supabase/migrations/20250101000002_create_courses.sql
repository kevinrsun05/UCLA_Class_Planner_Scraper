-- Create courses table (catalog data - timeless)
CREATE TABLE IF NOT EXISTS public.courses (
    id               bigserial PRIMARY KEY,
    subject_area_id  bigint NOT NULL REFERENCES public.subject_areas(id) ON DELETE CASCADE,
    number           text NOT NULL,
    title            text NOT NULL,
    description      text,
    units            text,
    requisites_text  text,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (subject_area_id, number)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_courses_subject_area ON public.courses(subject_area_id);
CREATE INDEX IF NOT EXISTS idx_courses_number ON public.courses(subject_area_id, number);

-- Enable Row Level Security
ALTER TABLE public.courses ENABLE ROW LEVEL SECURITY;

-- RLS Policies: Public read access, service role can modify
CREATE POLICY "Public read access for courses"
    ON public.courses
    FOR SELECT
    USING (true);

CREATE POLICY "Service role full access to courses"
    ON public.courses
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);
