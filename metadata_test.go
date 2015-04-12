// contains the tests for the metadata tests

package gnosis

import (
	"github.com/stretchr/testify/assert"
	//"os"
	//"path"
	"testing"
)

func stringKeyExistsInMap(itemMap map[string]bool, key string) bool {
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

func TestCheckMatch(t *testing.T) {
	pdata := new(PageMetadata)

	metadataLine := []byte("topic = a")
	metadataMatch := []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["a"], "should have been able to add element a to the thingy")

	metadataLine = []byte("topic= b")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["b"], "should have been able to add element b to the thingy")

	metadataLine = []byte("topic=c")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["c"], "should have been able to add element c to the thingy")

	metadataLine = []byte("topic=d e f")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["d-e-f"], "should have been able to add element d e f to the thingy")

	metadataLine = []byte("topic=g  h")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["g-h"], "should have been able to add element g-h to the thingy")

	metadataLine = []byte("topic:i")
	metadataMatch = []byte("topic")
	pdata.checkMatch(metadataLine, metadataMatch, &pdata.Topics)
	assert.True(t, pdata.Topics["i"], "should have been able to add element i to the thingy")
}

func TestProcessMetadata(t *testing.T) {
	pdata := new(PageMetadata)

	metaDataLine := []byte("topic = a")
	pdata.processMetadata(metaDataLine)
	assert.True(t, pdata.Topics["a"], "topic a should have been added")

	metaDataLine = []byte("keyword=b")
	pdata.processMetadata(metaDataLine)
	assert.True(t, pdata.Keywords["b"], "keyword b should have been added")
}
