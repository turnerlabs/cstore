package catalog

import (
	"testing"
)

func TestWhenAnyTagIsFoundReturnFile(t *testing.T) {
	// arrange
	files := map[string]File{
		"one": File{
			Tags: []string{"dev", "missing"},
		},
		"two": File{
			Tags: []string{"qa", "missing"},
		},
		"three": File{
			Tags: []string{"dev", "qa"},
		},
		"four": File{
			Tags: []string{"missing", "missing"},
		},
		"five": File{
			Tags: []string{},
		},
	}

	tags := []string{"dev", "qa"}

	// act
	results := keepFilesWithTags(files, tags, false)

	// assert
	for _, expected := range []string{"one", "two", "three"} {
		if _, found := results[expected]; !found {
			t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", expected)
		}
	}

	for _, notExpected := range []string{"four", "five"} {
		if _, found := results[notExpected]; found {
			t.Errorf("\nEXPECTED: file missing \nACTUAL: %s", notExpected)
		}
	}
}

func TestWhenAllTagsAreFoundReturnFile(t *testing.T) {
	// arrange
	files := map[string]File{
		"one": File{
			Tags: []string{"dev", "other"},
		},
		"two": File{
			Tags: []string{"qa", "other"},
		},
		"three": File{
			Tags: []string{"dev", "qa", "other"},
		},
		"four": File{
			Tags: []string{"other", "other"},
		},
		"five": File{
			Tags: []string{},
		},
	}

	tags := []string{"dev", "other"}

	// act
	results := keepFilesWithTags(files, tags, true)

	// assert
	for _, expected := range []string{"one", "three"} {
		if _, found := results[expected]; !found {
			t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", expected)
		}
	}

	for _, notExpected := range []string{"two", "four", "five"} {
		if _, found := results[notExpected]; found {
			t.Errorf("\nEXPECTED: file missing \nACTUAL: %s", notExpected)
		}
	}
}

func TestReferenceFilesWithValidTagsShouldMatch(t *testing.T) {
	// arrange
	files := map[string]File{
		"one": File{
			IsRef: true,
			Tags:  []string{"dev", "other"},
		},
		"two": File{
			IsRef: true,
			Tags:  []string{"dev"},
		},
	}

	tags := []string{"other"}

	// act
	results := keepFilesWithTags(files, tags, true)

	// assert
	for _, expected := range []string{"one"} {
		if _, found := results[expected]; !found {
			t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", expected)
		}
	}

	for _, notExpected := range []string{"two"} {
		if _, found := results[notExpected]; found {
			t.Errorf("\nEXPECTED: file missing \nACTUAL: %s", notExpected)
		}
	}
}

func TestReferenceFilesWithoutTagsShouldMatch(t *testing.T) {
	// arrange
	files := map[string]File{
		"one": File{
			IsRef: true,
			Tags:  []string{"dev", "other"},
		},
		"two": File{
			IsRef: true,
			Tags:  []string{},
		},
	}

	tags := []string{"test"}

	// act
	results := keepFilesWithTags(files, tags, true)

	// assert
	for _, expected := range []string{"two"} {
		if _, found := results[expected]; !found {
			t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", expected)
		}
	}

	for _, notExpected := range []string{"one"} {
		if _, found := results[notExpected]; found {
			t.Errorf("\nEXPECTED: file missing \nACTUAL: %s", notExpected)
		}
	}
}

func TestWhenNoTagsArePassedAllFilesShouldMatch(t *testing.T) {
	// arrange
	files := map[string]File{
		"one": File{
			IsRef: true,
			Tags:  []string{"dev", "other"},
		},
		"two": File{
			Tags: []string{},
		},
		"three": File{
			Tags: []string{"dev", "other"},
		},
	}

	// act
	results := keepFilesWithTags(files, []string{}, true)

	// assert
	for _, expected := range []string{"one", "two", "three"} {
		if _, found := results[expected]; !found {
			t.Errorf("\nEXPECTED: %s \nACTUAL: file missing", expected)
		}
	}
}
