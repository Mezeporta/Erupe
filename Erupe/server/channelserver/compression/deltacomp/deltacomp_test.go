package deltacomp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/Solenataris/Erupe/server/channelserver/compression/nullcomp"
)

var tests = []struct {
	before  string
	patches []string
	after   string
}{
	{
		"hunternavi_0_before.bin",
		[]string{
			"hunternavi_0_patch_0.bin",
			"hunternavi_0_patch_1.bin",
		},
		"hunternavi_0_after.bin",
	},
	{
		// From "Character Progression 1 Creation-NPCs-Tours"
		"hunternavi_1_before.bin",
		[]string{
			"hunternavi_1_patch_0.bin",
			"hunternavi_1_patch_1.bin",
			"hunternavi_1_patch_2.bin",
			"hunternavi_1_patch_3.bin",
			"hunternavi_1_patch_4.bin",
			"hunternavi_1_patch_5.bin",
			"hunternavi_1_patch_6.bin",
			"hunternavi_1_patch_7.bin",
			"hunternavi_1_patch_8.bin",
			"hunternavi_1_patch_9.bin",
			"hunternavi_1_patch_10.bin",
			"hunternavi_1_patch_11.bin",
			"hunternavi_1_patch_12.bin",
			"hunternavi_1_patch_13.bin",
			"hunternavi_1_patch_14.bin",
			"hunternavi_1_patch_15.bin",
			"hunternavi_1_patch_16.bin",
			"hunternavi_1_patch_17.bin",
			"hunternavi_1_patch_18.bin",
			"hunternavi_1_patch_19.bin",
			"hunternavi_1_patch_20.bin",
			"hunternavi_1_patch_21.bin",
			"hunternavi_1_patch_22.bin",
			"hunternavi_1_patch_23.bin",
			"hunternavi_1_patch_24.bin",
		},
		"hunternavi_1_after.bin",
	},
	{
		// From "Progress Gogo GRP Grind 9 and Armor Upgrades and Partner Equip and Lost Cat and Manager talk and Pugi Order"
		// Not really sure this one counts as a valid test as the input and output are exactly the same. The patches cancel each other out.
		"platedata_0_before.bin",
		[]string{
			"platedata_0_patch_0.bin",
			"platedata_0_patch_1.bin",
		},
		"platedata_0_after.bin",
	},
}

func readTestDataFile(filename string) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("./test_data/%s", filename))
	if err != nil {
		panic(err)
	}
	return data
}

func TestDeltaPatch(t *testing.T) {
	for k, tt := range tests {
		testname := fmt.Sprintf("delta_patch_test_%d", k)
		t.Run(testname, func(t *testing.T) {
			// Load the test binary data.
			beforeData, err := nullcomp.Decompress(readTestDataFile(tt.before))
			if err != nil {
				t.Error(err)
			}

			var patches [][]byte
			for _, patchName := range tt.patches {
				patchData := readTestDataFile(patchName)
				patches = append(patches, patchData)
			}

			afterData, err := nullcomp.Decompress(readTestDataFile(tt.after))
			if err != nil {
				t.Error(err)
			}

			// Now actually test calling ApplyDataDiff.
			data := beforeData

			// Apply the patches in order.
			for i, patch := range patches {
				fmt.Println("patch index: ", i)
				data = ApplyDataDiff(patch, data)
			}

			if !bytes.Equal(data, afterData) {
				t.Errorf("got out\n\t%s\nwant\n\t%s", hex.Dump(data), hex.Dump(afterData))
			}
		})
	}
}
