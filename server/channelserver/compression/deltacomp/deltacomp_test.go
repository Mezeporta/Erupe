package deltacomp

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"erupe-ce/server/channelserver/compression/nullcomp"
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
	data, err := os.ReadFile(fmt.Sprintf("./test_data/%s", filename))
	if err != nil {
		panic(err)
	}
	return data
}

func TestApplyDataDiffWithLimit_BoundsCheck(t *testing.T) {
	// Base data: 10 bytes
	baseData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}

	// Build a patch that tries to write at offset 8 with 5 different bytes,
	// which would extend to offset 13 (beyond 10-byte base).
	// Format: matchCount=9 (first is +1), differentCount=6 (is -1 = 5 bytes)
	diff := []byte{
		0x09,                         // matchCount (first is +1, so offset becomes -1+9=8)
		0x06,                         // differentCount (6-1=5 different bytes)
		0xAA, 0xBB, 0xCC, 0xDD, 0xEE, // 5 patch bytes
	}

	t.Run("within_limit", func(t *testing.T) {
		// Limit of 20 allows the growth
		result, err := ApplyDataDiffWithLimit(diff, baseData, 20)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) < 13 {
			t.Errorf("expected result length >= 13, got %d", len(result))
		}
	})

	t.Run("exceeds_limit", func(t *testing.T) {
		// Limit of 10 doesn't allow writing past the base
		_, err := ApplyDataDiffWithLimit(diff, baseData, 10)
		if err == nil {
			t.Error("expected error for write past limit, got none")
		}
	})

	t.Run("no_limit", func(t *testing.T) {
		// maxOutput=0 means no limit (backwards compatible)
		result, err := ApplyDataDiffWithLimit(diff, baseData, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) < 13 {
			t.Errorf("expected result length >= 13, got %d", len(result))
		}
	})
}

func TestApplyDataDiffWithLimit_TruncatedPatch(t *testing.T) {
	baseData := []byte{0x01, 0x02, 0x03, 0x04}

	// Patch claims 3 different bytes but only provides 1
	diff := []byte{
		0x02, // matchCount (offset = -1+2 = 1)
		0x04, // differentCount (4-1=3 different bytes)
		0xAA, // only 1 byte provided (missing 2)
	}

	_, err := ApplyDataDiffWithLimit(diff, baseData, 100)
	if err == nil {
		t.Error("expected error for truncated patch, got none")
	}
}

func TestApplyDataDiff_ReturnsOriginalOnError(t *testing.T) {
	baseData := []byte{0x01, 0x02, 0x03, 0x04}

	// Truncated patch
	diff := []byte{
		0x02,
		0x04,
		0xAA, // only 1 of 3 expected bytes
	}

	result := ApplyDataDiff(diff, baseData)
	// On error, ApplyDataDiff should return the original data unchanged
	if !bytes.Equal(result, baseData) {
		t.Errorf("expected original data on error, got %v", result)
	}
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
