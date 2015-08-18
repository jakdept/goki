
package gnosis

import (
	"github.com/stretchr/testify/assert"
	// "os"
	// "path"
	"testing"
	// "time"
)

func TestGetURIPath(t *testing.T) {
	var tests = []struct{
		input string
		trim string
		add string
		output string
	}{
		{"/wiki.md", "/", "/var/www/", "/var/www/wiki.md"},
		{"/wiki/page.md", "/wiki/", "/var/www/", "/var/www/page.md"},
		{"/wiki/page.md", "/wiki/", "/", "/page.md"},
		{"abcdef", "abc", "xyz", "xyzdef"},
	}

	for _, testSet := range tests {
		assert.Equal(t, testSet.output, getURIPath(testSet.input, testSet.trim, testSet.add),
			"[%q] trimmed [%q] added [%q] but got the wrong result",
			testSet.input, testSet.trim, testSet.add)
	}
}

// func TestCleanupMarkdown(t *testing.T) {
	
// }