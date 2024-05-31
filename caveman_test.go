package caveman

import "testing"

func TestStartup(t *testing.T) {
	t.Log("Test Startup")
	c := New()

	err := c.Start()
	if err != nil {
		t.Error(err)
	}
}
