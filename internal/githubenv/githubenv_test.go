package githubenv

import "testing"

func TestWrite(t *testing.T) {
	var buf writerRecorder
	err := write(&buf, map[string]string{
		"DRONE_COMMIT_MESSAGE": "line one\nline two",
		"DRONE_BRANCH":         "main",
	})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !containsInOrder(out, "DRONE_BRANCH<<", "main", "DRONE_COMMIT_MESSAGE<<", "line one\nline two") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

type writerRecorder struct{ data []byte }

func (w *writerRecorder) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}
func (w *writerRecorder) String() string { return string(w.data) }

func containsInOrder(s string, parts ...string) bool {
	for _, p := range parts {
		idx := indexOf(s, p)
		if idx < 0 {
			return false
		}
		s = s[idx+len(p):]
	}
	return true
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
