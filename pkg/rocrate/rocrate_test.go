package rocrate

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/go-test/deep"
)

// TestContexts ensures that we can open an empty JSON-LD as expected
// and reason about it.
func TestContexts(t *testing.T) {

	expectedContext := "https://w3id.org/ro/crate/1.1/context"

	simpleContextRes := bytes.NewBuffer(simpleContext)
	res, err := ProcessMetadataStream(simpleContextRes)

	if err != nil {
		t.Errorf("error processing simpleContext: %s", err)
	}
	if res.Context() != expectedContext {
		t.Errorf("context wasn't read successfully, got: '%s'", res.Context())
	}
	if len(res.Graph) != 0 {
		t.Errorf("expecting empty graph, got graph len: '%d'", len(res.Graph))
	}

	complexContextRes := bytes.NewBuffer(complexContext)
	res, err = ProcessMetadataStream(complexContextRes)

	if err != nil {
		t.Errorf("error processing complexContext: %s", err)
	}
	if res.Context() != expectedContext {
		t.Errorf("context wasn't read successfully, got: '%s'", res.Context())
	}
	if len(res.Graph) != 0 {
		t.Errorf("expecting empty graph, got graph len: '%d'", len(res.Graph))
	}
}

type stringTest struct {
	label    string
	testData []byte
	compare1 []string
	compare2 []string
}

var stringTests []stringTest = []stringTest{
	stringTest{
		"name",
		nameTest,
		[]string{"name one", "name two"},
		[]string{"name one"},
	},
	stringTest{
		"keyword",
		keywordTest,
		[]string{"kw1", "kw2"},
		[]string{"kw1"},
	},
	stringTest{
		"type",
		typeTest,
		[]string{"Person", "Artist"},
		[]string{"Person"},
	},
}

// getStringSliceValue gives us some way of dynamically accessing
// attributes to avoid a decent amount of code replication.
func getStringSliceValue(data graph, label string) []string {
	switch label {
	case "name":
		return data.Name.Value()
	case "keyword":
		return data.Keywords.Value()
	case "type":
		return data.Type.Value()
	}
	return []string{""}
}

// TestStringVariants tests our ability to decode single-value strings
// or slices. We convert the single-value to a slice so we expect a
// slice array at all times.
func TestStringVariants(t *testing.T) {
	for _, test := range stringTests {
		variantTest := bytes.NewBuffer(test.testData)
		res, err := ProcessMetadataStream(variantTest)

		if err != nil {
			t.Errorf("%s: cannot process input ('%s')", test.label, err)
		}
		if len(res.Graph) != 2 {
			fmt.Printf("test data is incorrect length: '%d'", len(res.Graph))
		}
		testID := res.Graph[0].ID
		if testID != "test1" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value := getStringSliceValue(res.Graph[0], test.label)
		if !slices.Equal(value, test.compare1) {
			t.Errorf(
				"%s: string variant: '%v' result doesn't match expected: '%v'",
				fmt.Sprintf("%s test 1", test.label),
				value,
				test.compare1,
			)
		}
		testID = res.Graph[1].ID
		if testID != "test2" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value = getStringSliceValue(res.Graph[1], test.label)
		if !slices.Equal(value, test.compare2) {
			t.Errorf(
				"%s: string variant: '%v' result doesn't match expected: '%v'",
				fmt.Sprintf("%s test 2", test.label),
				value,
				test.compare2,
			)
		}
	}
}

type nodeTest struct {
	label    string
	testData []byte
	compare1 []nodeIdentifier
	compare2 []nodeIdentifier
}

var nodeTests []nodeTest = []nodeTest{
	nodeTest{
		"author",
		authorTest,
		[]nodeIdentifier{
			nodeIdentifier{"https://orcid.org/1234-0003-1974-0000"},
			nodeIdentifier{"#Yann"},
			nodeIdentifier{"#Organization-SMUC"},
		},
		[]nodeIdentifier{
			nodeIdentifier{"https://orcid.org/0000-0003-1974-1234"},
		},
	},
	nodeTest{
		"hasPart",
		hasPartTest,
		[]nodeIdentifier{
			nodeIdentifier{"part1"},
		},
		[]nodeIdentifier{
			nodeIdentifier{"part1"},
			nodeIdentifier{"part2"},
			nodeIdentifier{"part3"},
		},
	},
	nodeTest{
		"license",
		licenseTest,
		[]nodeIdentifier{
			nodeIdentifier{"https://spdx.org/licenses/license1"},
		},
		[]nodeIdentifier{
			nodeIdentifier{"http://spdx.org/licenses/license2"},
		},
	},
}

// getNodeIdentifierSliceValue allows us to get values more dynamically
// from the nodeIdentifier tests.
func getNodeIdentifierSliceValue(data graph, label string) []nodeIdentifier {
	switch label {
	case "author":
		return data.Author.Value()
	case "license":
		return data.License.Value()
	case "hasPart":
		return data.HasPart.Value()
	}
	return []nodeIdentifier{}
}

// TestNodeIdentifierVariants tests the conversion of single-values to
// a slice of nodeIdentifiers.
func TestNodeIdentifierVariants(t *testing.T) {
	for _, test := range nodeTests {
		variantTest := bytes.NewBuffer(test.testData)
		res, err := ProcessMetadataStream(variantTest)

		if err != nil {
			t.Errorf("%s: cannot process input ('%s')", test.label, err)
		}
		if len(res.Graph) != 2 {
			fmt.Printf(
				"%s: test data is incorrect length: '%d'",
				test.label,
				len(res.Graph),
			)
		}
		testID := res.Graph[0].ID
		if testID != "test1" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value := getNodeIdentifierSliceValue(res.Graph[0], test.label)
		for idx, v := range value {
			if v.ID != test.compare1[idx].ID {
				t.Errorf(
					"%s: string variant: '%v' result doesn't match expected: '%v'",
					fmt.Sprintf("%s test 1", test.label),
					v,
					test.compare1[idx],
				)
			}
		}
		testID = res.Graph[1].ID
		if testID != "test2" {
			t.Errorf("test ID is incorrect: %s", testID)
		}
		value = getNodeIdentifierSliceValue(res.Graph[1], test.label)
		for idx, v := range value {
			if v.ID != test.compare2[idx].ID {
				t.Errorf(
					"%s: string variant: '%v' result doesn't match expected: '%v'",
					fmt.Sprintf("%s test 2", test.label),
					v,
					test.compare2[idx],
				)
			}
		}
	}
}

// TestNewSummary ensures new summary is as safe as possible.
func TestNewSummary(t *testing.T) {
	summary := newSummary()
	structType := reflect.TypeOf(summary)
	structVal := reflect.ValueOf(summary)
	fieldNum := structVal.NumField()
	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)
		if fmt.Sprintf("%s", field.Type()) == "string" {
			continue
		}
		fieldName := structType.Field(i).Name
		isSet := field.IsValid() && !field.IsZero()
		if !isSet {
			t.Errorf("summary constructor isn't setting: %s", fieldName)
		}
	}
}

// TestNewGocflSummary ensures new gocfl summary is as safe as possible.
func TestNewGocflSummary(t *testing.T) {
	summary := newGocflSummary()
	structType := reflect.TypeOf(summary)
	structVal := reflect.ValueOf(summary)
	fieldNum := structVal.NumField()
	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)
		if fmt.Sprintf("%s", field.Type()) == "string" {
			continue
		}
		fieldName := structType.Field(i).Name
		isSet := field.IsValid() && !field.IsZero()
		if !isSet {
			t.Errorf("summary constructor isn't setting: %s", fieldName)
		}
	}
}

type metadataTest struct {
	label        string
	testData     []byte
	summary      RocrateSummary
	gocflSummary GocflSummary
}

var metadataTests []metadataTest = []metadataTest{
	metadataTest{
		"empty",
		emptyCrate,
		RocrateSummary{
			// ID
			"",
			// Name
			[]string{},
			// Type
			[]string{},
			// Description
			[]string{},
			// DatePublished
			"",
			// Author
			[]string{},
			// License
			"",
			// HasPart
			[]string{},
			// ContentURL
			[]string{},
			// Keywords
			[]string{},
			// Publisher
			[]string{},
			// About
			"./",
		},
		GocflSummary{
			// signature
			"",
			// title
			"",
			// description
			"",
			// created
			"",
			// sets
			"",
			// keywords
			"",
			// licenses
			"",
			// provided by caller.
			"",
			"",
			"",
			"",
			"",
		},
	},
	metadataTest{
		"afternoon drinks",
		afternoonDrinks,
		RocrateSummary{
			// ID
			"./RC0E772B3021E7E40C2BBDE657",
			// Name
			[]string{"A study of my afternoon drinks"},
			// Type
			[]string{"Dataset"},
			// Description
			[]string{
				"A study of my afternoon consumption one week in 2018",
				"Exported from Dataverse",
			},
			// DatePublished
			"2018",
			// Author
			[]string{"#Ross_Spencer-1"},
			// License
			"https://spdx.org/licenses/CC0-1.0.html",
			// HasPart
			[]string{
				"metadata/agents.json",
				"metadata/dataset.json",
				"Drinkscitation-endnote.xml",
				"Drinks.tab",
				"Drinks.csv",
				"Drinkscitation-ris.ris",
				"Drinkscitation-bib.bib",
				"Drinks.RData",
				"Drinks-ddi.xml",
			},
			// ContentURL
			nil,
			// Keywords
			[]string{
				"dataverse",
				"study",
				"observational-study",
			},
			// Publisher
			[]string{
				"https://ror.org/02s6k3f65",
			},
			// About
			"./RC0E772B3021E7E40C2BBDE657",
		},
		GocflSummary{
			// signature
			"./RC0E772B3021E7E40C2BBDE657",
			// title
			"A study of my afternoon drinks",
			// description
			"A study of my afternoon consumption one week in 2018; Exported from Dataverse",
			// created
			"2018",
			// sets
			"Dataset",
			// keywords
			"dataverse; study; observational-study",
			// licenses
			"https://spdx.org/licenses/CC0-1.0.html",
			// provided by caller.
			"",
			"",
			"",
			"",
			"",
		},
	},
	metadataTest{
		"carpentries",
		carpentriesCrate,
		RocrateSummary{
			// ID
			"./",
			// Name
			[]string{
				"Example dataset for RO-Crate specification",
			},
			// Type
			[]string{
				"Dataset",
				"LearningResource",
			},
			// Description
			[]string{
				"Official rainfall readings for Katoomba, NSW 2022, Australia",
			},
			// DatePublished
			"2023-05-22T12:03:00+0100",
			// Author
			[]string{
				"https://orcid.org/0000-0002-1825-0097",
			},
			// License
			"http://spdx.org/licenses/CC0-1.0",
			// HasPart
			[]string{
				"data.csv",
			},
			// ContentURL
			[]string{},
			// Keywords
			[]string{},
			// Publisher
			[]string{
				"https://ror.org/05gq02987",
			},
			// About
			"./",
		},
		GocflSummary{
			// signature
			"./",
			// title
			"Example dataset for RO-Crate specification",
			// description
			"Official rainfall readings for Katoomba, NSW 2022, Australia",
			// created
			"2023-05-22T12:03:00+0100",
			// sets
			"Dataset; LearningResource",
			// keywords
			"",
			// licenses
			"http://spdx.org/licenses/CC0-1.0",
			// provided by caller.
			"",
			"",
			"",
			"",
			"",
		},
	},
	metadataTest{
		"spec",
		specCrate,
		RocrateSummary{
			// ID
			"./",
			// Name
			[]string{
				"Example RO-Crate",
			},
			// Type
			[]string{
				"Dataset",
			},
			// Description
			[]string{
				"The RO-Crate Root Data Entity",
			},
			// DatePublished
			"",
			// Author
			[]string{},
			// License
			"",
			// HasPart
			[]string{
				"data1.txt",
				"data2.txt",
			},
			// ContentURL
			[]string{},
			// Keywords
			[]string{},
			// Publisher
			[]string{},
			// About
			"./",
		},
		GocflSummary{
			// signature
			"./",
			// title
			"Example RO-Crate",
			// description
			"The RO-Crate Root Data Entity",
			// created
			"",
			// sets
			"Dataset",
			// keywords
			"",
			// licenses
			"",
			// provided by caller.
			"",
			"",
			"",
			"",
			"",
		},
	},
	metadataTest{
		"galaxy",
		galaxyCrate,
		RocrateSummary{
			// ID
			"./",
			// Name
			[]string{
				"Demo Crate",
			},
			// Type
			[]string{
				"Dataset",
				"LearningResource",
			},
			// Description
			[]string{
				"a demo crate for Galaxy training",
			},
			// DatePublished
			"2024-03-08",
			// Author
			[]string{
				"https://orcid.org/0000-0000-0000-0000",
			},
			// License
			"https://spdx.org/licenses/CC-BY-NC-SA-4.0.html",
			// HasPart
			[]string{
				"data.csv",
			},
			// ContentURL
			[]string{
				"http://example.com/resource/data",
			},
			// Keywords
			[]string{},
			// Publisher
			[]string{
				"https://ror.org/0abcdef00",
			},
			// About
			"./",
		},
		GocflSummary{
			// signature
			"./",
			// title
			"Demo Crate",
			// description
			"a demo crate for Galaxy training",
			// created
			"2024-03-08",
			// sets
			"Dataset; LearningResource",
			// keywords
			"",
			// licenses
			"https://spdx.org/licenses/CC-BY-NC-SA-4.0.html",
			// provided by caller.
			"",
			"",
			"",
			"",
			"",
		},
	},
}

// TestMetadata provides more generic testing of the test data within
// this package.
func TestMetadata(t *testing.T) {
	for _, test := range metadataTests {
		variantTest := bytes.NewBuffer(test.testData)
		processed, err := ProcessMetadataStream(variantTest)
		if err != nil {
			t.Errorf("%s: cannot process input ('%s')", test.label, err)
		}
		res, _ := processed.Summary()
		if diff := deep.Equal(res, test.summary); diff != nil {
			t.Errorf("%s summary metadata doesn't match: %s", test.label, diff)
		}
	}
}

// TestGOCFLMetadata provides more generic testing the summary output
// of the GOCFL struct.
func TestGOCFLMetadata(t *testing.T) {
	for _, test := range metadataTests {
		variantTest := bytes.NewBuffer(test.testData)
		processed, _ := ProcessMetadataStream(variantTest)
		currentTime = func(bool) string {
			// Mock getTime function.
			return ""
		}
		res, _ := processed.GOCFLSummary()
		if diff := deep.Equal(res, test.gocflSummary); diff != nil {
			t.Errorf("%s gocfl metadata doesn't match: %s", test.label, diff)
		}
	}
}
