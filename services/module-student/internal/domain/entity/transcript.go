package entity

import "time"

// TranscriptEntry is one academic result line for a student.
type TranscriptEntry struct {
	EnrollmentID     string
	SemesterID       string
	SubjectID        string
	SubjectCode      string
	SubjectName      string
	Credits          int32
	Status           string
	GradeNumeric     float64
	GradeLetter      string
	GradedAt         *time.Time
}

// StudentTranscript is the assembled transcript response model.
type StudentTranscript struct {
	Student        *Student
	Entries        []TranscriptEntry
	GPA            float64
	TotalCredits   int32
	PassedCredits  int32
}
