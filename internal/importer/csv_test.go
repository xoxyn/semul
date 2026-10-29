package importer

import (
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	got := Read()

	want := []string{
		"first_name", "last_name", "username",
		"Rob", "Pike", "rob",
		"Ken", "Thompson", "ken",
		"Robert", "Griesemer", "gri",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Read() = %v, want %v", got, want)
	}
}
