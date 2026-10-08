package consentservice

import (
	"testing"
	"time"
)

func TestDecideCourseAccess(t *testing.T) {
	deadline := time.Date(2026, time.October, 15, 17, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		granted    bool
		wantState  string
		wantAccess bool
	}{
		{name: "grant enables delivery deadline and reporting", granted: true, wantState: "consent_granted", wantAccess: true},
		{name: "revocation removes delivery deadline and reporting", granted: false, wantState: "consent_revoked", wantAccess: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideCourseAccess(CourseRequest{
				UserID: "learner-42", CourseID: "math-201", Deadline: deadline, ConsentGranted: tt.granted,
			})
			if got.State != tt.wantState {
				t.Fatalf("state = %q, want %q", got.State, tt.wantState)
			}
			if got.DeliverCourse != tt.wantAccess || got.DeadlineVisible != tt.wantAccess || got.IncludeInEducatorReport != tt.wantAccess {
				t.Fatalf("access flags = (%t, %t, %t), want all %t", got.DeliverCourse, got.DeadlineVisible, got.IncludeInEducatorReport, tt.wantAccess)
			}
			if got.Deadline != deadline {
				t.Fatalf("deadline = %s, want %s", got.Deadline, deadline)
			}
		})
	}
}
