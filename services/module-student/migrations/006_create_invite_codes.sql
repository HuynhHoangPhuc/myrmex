-- +goose Up
CREATE TABLE IF NOT EXISTS student.invite_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hex of the plaintext code
    student_id UUID NOT NULL REFERENCES student.students(id),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    used_by_user_id UUID,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_codes_hash ON student.invite_codes(code_hash);
CREATE INDEX IF NOT EXISTS idx_invite_codes_student ON student.invite_codes(student_id);

-- +goose Down
DROP TABLE IF EXISTS student.invite_codes;
