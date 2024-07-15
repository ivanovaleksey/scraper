package scraper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_combinePath(t *testing.T) {
	testCases := []struct {
		basePath string
		relPath  string
		expected string
	}{
		{
			basePath: "catalogue/",
			relPath:  "a-light-in-the-attic_1000/index.html",
			expected: "catalogue/a-light-in-the-attic_1000/index.html",
		},
		{
			basePath: "index.html",
			relPath:  "a-light-in-the-attic_1000/index.html",
			expected: "a-light-in-the-attic_1000/index.html",
		},
		{
			basePath: "catalogue/category/books_1/index.html",
			relPath:  "../../a-light-in-the-attic_1000/index.html",
			expected: "catalogue/a-light-in-the-attic_1000/index.html",
		},
		{
			basePath: "catalogue/category/books/travel_2/index.html",
			relPath:  "../../../its-only-the-himalayas_981/index.html",
			expected: "catalogue/its-only-the-himalayas_981/index.html",
		},
		{
			basePath: "catalogue/category/books/travel_2/index.html",
			relPath:  "../../../../media/cache/27/a5/27a53d0bb95bdd88288eaf66c9230d7e.jpg",
			expected: "media/cache/27/a5/27a53d0bb95bdd88288eaf66c9230d7e.jpg",
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test case %d", i+1), func(t *testing.T) {
			actual := combinePath(testCase.basePath, testCase.relPath)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
