-- Create subject_areas table
CREATE TABLE IF NOT EXISTS public.subject_areas (
    id         bigserial PRIMARY KEY,
    code       text NOT NULL UNIQUE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- Create index on code for faster lookups
CREATE INDEX IF NOT EXISTS idx_subject_areas_code ON public.subject_areas(code);

-- Enable Row Level Security
ALTER TABLE public.subject_areas ENABLE ROW LEVEL SECURITY;

-- RLS Policies: Public read access, service role can modify
CREATE POLICY "Public read access for subject_areas"
    ON public.subject_areas
    FOR SELECT
    USING (true);

CREATE POLICY "Service role full access to subject_areas"
    ON public.subject_areas
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);
