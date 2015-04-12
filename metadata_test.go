// contains the tests for the metadata tests

package gnosis

import (
	"github.com/stretchr/testify/assert"
	//"os"
	//"path"
	"testing"
)

func existsInMap(itemMap map[string]bool, key string) bool {
	defer func() bool {
		r := recover()
		return r == nil
	}()
	junk := itemMap[key]
	return junk == junk
}

func TestLineIsTitle(t *testing.T) {
	pdata := new(PageMetadata)

	// test the most normal topic line I'd expect
	titleLine := []byte("=======")
	assert.True(t, pdata.lineIsTitle(titleLine), "the most basic topic line just failed")

	titleLine = []byte("=")
	assert.True(t, pdata.lineIsTitle(titleLine), "one = should be enough")

	titleLine = []byte("   ======")
	assert.True(t, pdata.lineIsTitle(titleLine), "any spaces before the heading portion should not cause failure")

	titleLine = []byte("			======")
	assert.True(t, pdata.lineIsTitle(titleLine), "tabs before the heading portion should not cause failure")

	titleLine = []byte("=======     ")
	assert.True(t, pdata.lineIsTitle(titleLine), "spaces after the heading portion should not cause failure")

	titleLine = []byte("=======			")
	assert.True(t, pdata.lineIsTitle(titleLine), "tabs after the heading portion should not cause failure")

	titleLine = []byte("=======\n")
	assert.True(t, pdata.lineIsTitle(titleLine), "a newline after the heading portion should not cause failure")

	titleLine = []byte("===== ===")
	assert.False(t, pdata.lineIsTitle(titleLine), "the underlining has to be continous - no spaces - so this should have failed")

	titleLine = []byte("====	=====")
	assert.False(t, pdata.lineIsTitle(titleLine), "the underlining has to be continous - no tabs - so this should have failed")
}

func TestProcessMetadata(t *testing.T) {
	pdata := new(PageMetadata)

	metaDataLine := []byte("topic=a")
	pdata.processMetadata(metaDataLine)
	assert.True(t, existsInMap(pdata.Topics, "a"), "the element I just tried to add should have been added")
}
