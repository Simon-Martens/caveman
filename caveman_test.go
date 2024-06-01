package caveman

import (
	"os"
	"strings"
	"testing"
)

func TestStartup(t *testing.T) {

	// copy os.Args
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)
	defer func() {
		// restore os.Args
		os.Args = originalArgs
	}()

	// change os.Args
	os.Args = os.Args[:1]
	os.Args = append(
		os.Args,
		"serve",
		"--dir=cm_test",
		"--dev",
	)

	app := New()

	if app == nil {
		t.Fatal("Expected initialized PocketBase instance, got nil")
	}

	if app.RootCmd == nil {
		t.Fatal("Expected RootCmd to be initialized, got nil")
	}

	if app.DataDir() != "test_dir" {
		t.Fatalf("Expected app.DataDir() %q, got %q", "test_dir", app.DataDir())
	}

	if app.IsDev() != true {
		t.Fatalf("Expected app.IsDev() %v, got %v", true, app.IsDev())
	}

	t.Log("Test Startup")

	err := app.Start()
	if err != nil {
		t.Error(err)
	}

	app.Logger().Info(strings.Join(os.Args, " "))
}
