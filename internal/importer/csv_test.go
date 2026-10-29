package importer

import (
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	input := `company,email,identified_problems
Acme,hello@acme.com,Manual workflows
Globex,info@globex.com,"Slow reporting, repetitive tasks"
`

	got, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d prospects, want 2", len(got))
	}

	if got[0].Company != "Acme" {
		t.Errorf("got company %q, want %q", got[0].Company, "Acme")
	}

	if got[1].IdentifiedProblem != "Slow reporting, repetitive tasks" {
		t.Errorf("unexpected problem: %q", got[1].IdentifiedProblem)
	}
}
