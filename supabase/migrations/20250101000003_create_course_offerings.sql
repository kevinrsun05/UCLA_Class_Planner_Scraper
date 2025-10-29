-- Create course_offerings table (term-specific data)
CREATE TABLE IF NOT EXISTS public.course_offerings (
    id               bigserial PRIMARY KEY,
    course_id        bigint NOT NULL REFERENCES public.courses(id) ON DELETE CASCADE,
    term             text NOT NULL,
    section          text,
    instructor       text,
    meeting_times    text,
    location         text,
    enrollment_status text,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (course_id, term, section)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_course_offerings_course ON public.course_offerings(course_id);
CREATE INDEX IF NOT EXISTS idx_course_offerings_term ON public.course_offerings(term);
CREATE INDEX IF NOT EXISTS idx_course_offerings_course_term ON public.course_offerings(course_id, term);

-- Enable Row Level Security
ALTER TABLE public.course_offerings ENABLE ROW LEVEL SECURITY;

-- RLS Policies: Public read access, service role can modify
CREATE POLICY "Public read access for course_offerings"
    ON public.course_offerings
    FOR SELECT
    USING (true);

CREATE POLICY "Service role full access to course_offerings"
    ON public.course_offerings
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);
