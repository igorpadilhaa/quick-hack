package main

import (
	"slices"
	"testing"
)

func TestPathSet(t *testing.T) {
	pathSet := PathSet{}

	if len(pathSet) != 0 {
		t.Errorf("new path set should be empty; got lenght %d", len(pathSet))
	}

	pathSet.AddAll([]string{"new path"})
	if len(pathSet) != 1 {
		t.Error("fail add new entry")
	}

	expected := []string{"new path"}
	if entries := pathSet.Entries(); !slices.Equal(expected, entries) {
		t.Errorf("fail to generate entries: got %q; expected %q", entries, expected)
	}
}

func TestDependencyResolve(t *testing.T) {
	type testCase struct {
		Name                    string
		TestAppName             string
		ExpectedDependencyCount int
	}

	testTable := []testCase{
		{
			Name:                    "single dependency resolval",
			TestAppName:             "java",
			ExpectedDependencyCount: 1,
		},
		{
			Name:                    "transitive dependency resolval",
			TestAppName:             "gradle",
			ExpectedDependencyCount: 2,
		},
		{
			Name:                    "cyclic dependency resolval",
			TestAppName:             "php",
			ExpectedDependencyCount: 2,
		},
	}

	catalog := AppCatalog{
		"java":     App{"/path/java", nil},
		"gradle":   App{"/path/gradle", []string{"java"}},
		"php":      App{"/path/php", []string{"composer"}},
		"composer": App{"/path/composer", []string{"php"}},
	}

	for _, testCase := range testTable {
		resolved, err := catalog.ResolveDependencies(testCase.TestAppName)
		t.Logf("Test case: %s", testCase.Name)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		expectedCount := testCase.ExpectedDependencyCount
		resolvedCount := len(resolved)

		if resolvedCount != expectedCount {
			t.Errorf("fail to resolve dependencies; expected count %d, got %d", expectedCount, resolvedCount)
		}

	}
}
