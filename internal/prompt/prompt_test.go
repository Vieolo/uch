package prompt

import "testing"

func TestSubstitute(t *testing.T) {
	got, err := Substitute("flutter build {{platform}}", map[string]string{"platform": "ios"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "flutter build ios" {
		t.Errorf("got %q, want %q", got, "flutter build ios")
	}
}

func TestSubstituteMultiple(t *testing.T) {
	got, err := Substitute("kubectl --context={{ctx}} get pods -n {{ns}}", map[string]string{
		"ctx": "prod",
		"ns":  "default",
	})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "kubectl --context=prod get pods -n default" {
		t.Errorf("got %q", got)
	}
}

func TestSubstituteUnresolved(t *testing.T) {
	_, err := Substitute("echo {{missing}}", map[string]string{"other": "x"})
	if err == nil {
		t.Fatal("expected error for unresolved placeholder")
	}
}

func TestSubstituteNoPlaceholders(t *testing.T) {
	got, err := Substitute("ls -la", nil)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "ls -la" {
		t.Errorf("got %q", got)
	}
}
