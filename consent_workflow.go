package consentservice

import "time"

const LearningRecordsCategory = "essential"

type CourseRequest struct {
	UserID         string    `json:"user_id"`
	CourseID       string    `json:"course_id"`
	Deadline       time.Time `json:"deadline"`
	Action         string    `json:"action"`
	Category       string    `json:"category"`
	RequestID      string    `json:"request_id"`
	ConsentGranted bool      `json:"-"`
}

type CourseDecision struct {
	UserID                  string    `json:"user_id"`
	CourseID                string    `json:"course_id"`
	Deadline                time.Time `json:"deadline"`
	DeliverCourse           bool      `json:"deliver_course"`
	DeadlineVisible         bool      `json:"deadline_visible"`
	IncludeInEducatorReport bool      `json:"include_in_educator_report"`
	State                   string    `json:"state"`
}

func DecideCourseAccess(in CourseRequest) CourseDecision {
	state := "consent_revoked"
	if in.ConsentGranted {
		state = "consent_granted"
	}
	return CourseDecision{
		UserID:                  in.UserID,
		CourseID:                in.CourseID,
		Deadline:                in.Deadline,
		DeliverCourse:           in.ConsentGranted,
		DeadlineVisible:         in.ConsentGranted,
		IncludeInEducatorReport: in.ConsentGranted,
		State:                   state,
	}
}
