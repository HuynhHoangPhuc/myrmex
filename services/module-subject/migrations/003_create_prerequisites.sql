-- +goose Up
CREATE TABLE subject.prerequisites (
    subject_id UUID NOT NULL REFERENCES subject.subjects(id) ON DELETE CASCADE,
    prerequisite_id UUID NOT NULL REFERENCES subject.subjects(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL DEFAULT 'hard' CHECK (type IN ('hard', 'soft')),
    priority INT NOT NULL DEFAULT 1 CHECK (priority BETWEEN 1 AND 5),
    PRIMARY KEY (subject_id, prerequisite_id),
    -- prevent self-referencing
    CONSTRAINT no_self_ref CHECK (subject_id != prerequisite_id)
);

CREATE INDEX idx_prerequisites_subject ON subject.prerequisites(subject_id);
CREATE INDEX idx_prerequisites_prereq ON subject.prerequisites(prerequisite_id);

-- +goose Down
DROP TABLE IF EXISTS subject.prerequisites;
