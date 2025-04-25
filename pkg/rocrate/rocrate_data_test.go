package rocrate

// simplecontext contains bare minimum context.
var simpleContext []byte = []byte(`
{
  "@context": "https://w3id.org/ro/crate/1.1/context",
  "@graph": [
  ]
}
`)

// complexcontext provides further information about the schema and
// vocabulary.
var complexContext []byte = []byte(`
{
  "@context": [
    "https://w3id.org/ro/crate/1.1/context",
    {
      "@vocab": "http://schema.org/"
    }
  ],
  "@graph": [
  ]
}
`)

// Variants of different values that can be single values or slices.

// nameTest provides a way of testing a single value name or a slice
// of names.
var nameTest []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
        {
            "@id": "test1",
            "name": [
                "name one",
                "name two"
            ]
        },
        {
            "@id": "test2",
            "name": "name one"
        }
    ]
}
`)

// keywordTest provides a way to test a single-value keyword or a
// slice of keywords.
var keywordTest []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
        {
            "@id": "test1",
            "keywords": [
                "kw1",
                "kw2"
            ]
        },
        {
            "@id": "test2",
            "keywords": "kw1"
        }
    ]
}
`)

// typeTest provides a way to test a single-value type or a slice of
// values for the same key.
var typeTest []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
        {
            "@id": "test1",
            "@type": [
                "Person",
                "Artist"
            ]
        },
        {
            "@id": "test2",
            "@type": "Person"
        }
    ]
}
`)

// authorTest provides a way to test a single-value author, or a slice
// of author values.
var authorTest []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
        {
            "@id": "test1",
            "author": [
                {
                    "@id": "https://orcid.org/1234-0003-1974-0000"
                },
                {
                    "@id": "#Yann"
                },
                {
                    "@id": "#Organization-SMUC"
                }
            ]
        },
		{
            "@id": "test2",
            "author": {
				"@id": "https://orcid.org/0000-0003-1974-1234"
            }
        }
    ]
}
`)

var hasPartTest []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
		{
            "@id": "test1",
            "hasPart": {
				"@id": "part1"
            }
        },
        {
            "@id": "test2",
            "hasPart": [
                {
                    "@id": "part1"
                },
                {
                    "@id": "part2"
                },
                {
                    "@id": "part3"
                }
            ]
        }
    ]
}
`)

var licenseTest []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
        {
            "@id": "test1",
            "license": "https://spdx.org/licenses/license1"
        },
        {
            "@id": "test2",
            "license": {
                "@id": "http://spdx.org/licenses/license2"
            }
        }
    ]
}
`)

// afternoonDrinks is an example RO-CRATE meta created from Dataverse.
var afternoonDrinks []byte = []byte(`
{
    "@context": [
        "https://w3id.org/ro/crate/1.1/context",
        {
            "@vocab": "http://schema.org/"
        }
    ],
    "@graph": [
        {
            "@id": "ro-crate-metadata.json",
            "@type": "CreativeWork",
            "identifier": "ro-crate-metadata.json",
            "about": {
                "@id": "./RC0E772B3021E7E40C2BBDE657"
            }
        },
        {
            "@id": "./RC0E772B3021E7E40C2BBDE657",
            "@type": "Dataset",
            "name": "A study of my afternoon drinks",
            "description": [
                "A study of my afternoon consumption one week in 2018",
                "Exported from Dataverse"
            ],
            "datePublished": "2018",
            "license": "https://spdx.org/licenses/CC0-1.0.html",
            "hasPart": [
                {
                    "@id": "metadata/agents.json"
                },
                {
                    "@id": "metadata/dataset.json"
                },
                {
                    "@id": "Drinkscitation-endnote.xml"
                },
                {
                    "@id": "Drinks.tab"
                },
                {
                    "@id": "Drinks.csv"
                },
                {
                    "@id": "Drinkscitation-ris.ris"
                },
                {
                    "@id": "Drinkscitation-bib.bib"
                },
                {
                    "@id": "Drinks.RData"
                },
                {
                    "@id": "Drinks-ddi.xml"
                }
            ],
            "contentUrl": [],
            "keywords": [
                "dataverse",
                "study",
                "observational-study"
            ],
            "author": {
                "@id": "#Ross_Spencer-1"
            },
            "publisher": {
                "@id": "https://ror.org/02s6k3f65"
            },
            "funder": {
                "@id": "https://ror.org/02s6k3f65"
            }
        },
        {
            "@id": "metadata/agents.json",
            "@type": "File"
        },
        {
            "@id": "metadata/dataset.json",
            "@type": "File"
        },
        {
            "@id": "Drinkscitation-endnote.xml",
            "@type": "File"
        },
        {
            "@id": "Drinks.tab",
            "@type": "File"
        },
        {
            "@id": "Drinks.csv",
            "@type": "File",
            "description": "Primary information recorded for the study",
            "datePublished": "2018",
            "contentLocation": {
                "@id": "#Toronto-1"
            }
        },
        {
            "@id": "Drinkscitation-ris.ris",
            "@type": "File"
        },
        {
            "@id": "Drinkscitation-bib.bib",
            "@type": "File"
        },
        {
            "@id": "Drinks.RData",
            "@type": "File"
        },
        {
            "@id": "Drinks-ddi.xml",
            "@type": "File"
        },
        {
            "@id": "#CSV_data_with_Dataverse_citation-1",
            "@type": "Dataset",
            "name": "CSV_data_with_Dataverse_citation-1",
            "license": "https://spdx.org/licenses/CC0-1.0.html"
        },
        {
            "@id": "#Ross_Spencer-1",
            "@type": "Person",
            "name": "Ross",
            "familyName": "Spencer",
            "givenName": "b33tk33p3r",
            "funder": {
                "@id": "https://ror.org/02s6k3f65"
            },
            "affiliation": {
                "@id": "https://ror.org/02s6k3f65"
            },
            "address": "101 example.com str.",
            "email": "ross@example.com",
            "identifier": "RS4FCB6A76D83F4B39B542EF5D"
        },
        {
            "@id": "https://ror.org/02s6k3f65",
            "@type": "Organization",
            "name": "University of Basel"
        },
        {
            "@id": "#Toronto-1",
            "@type": "Place",
            "name": "Toronto-1",
            "description": "A location somewhere in Canada",
            "keywords": [
                "dataverse",
                "csv",
                "behavioral analysis"
            ]
        }
    ]
}
`)

// carprentriesCrate conforms to the example crate created as part of
// the RO-CRATE carpentries tutorial.
var carpentriesCrate []byte = []byte(`
{
	"@context": "https://w3id.org/ro/crate/1.1/context",
	"@graph": [
	  {
		"@id": "ro-crate-metadata.json",
		"@type": "CreativeWork",
		"conformsTo": {"@id": "https://w3id.org/ro/crate/1.1"},
		"about": {"@id": "./"}
	  },
	  {
		"@id": "./",
		"@type": ["Dataset", "LearningResource"],
		"hasPart": [
		  {"@id": "data.csv"}
		],
		"name": "Example dataset for RO-Crate specification",
		"description": "Official rainfall readings for Katoomba, NSW 2022, Australia",
		"datePublished": "2023-05-22T12:03:00+0100",
		"license": {"@id": "http://spdx.org/licenses/CC0-1.0"},
		"author": { "@id": "https://orcid.org/0000-0002-1825-0097" },
		"publisher": {"@id": "https://ror.org/05gq02987"}
	  },
	  {
		"@id": "data.csv",
		"@type": "File",
		"name": "Rainfall Katoomba 2022-02",
		"description": "Rainfall data for Katoomba, NSW Australia February 2022",
		"encodingFormat": "text/csv",
		"license": {"@id": "https://creativecommons.org/licenses/by-nc-sa/4.0/"}
	  },
	  {
		"@id": "https://creativecommons.org/licenses/by-nc-sa/4.0/",
		"@type": "CreativeWork",
		"name": "CC BY-NC-SA 4.0 International",
		"description": "Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International"
	  },
	  {
		"@id": "http://spdx.org/licenses/CC0-1.0",
		"@type": "CreativeWork",
		"name": "CC0-1.0",
		"description": "Creative Commons Zero v1.0 Universal",
		"url": "https://creativecommons.org/publicdomain/zero/1.0/"
	  },
	  {
		"@id": "https://orcid.org/0000-0002-1825-0097",
		"@type": "Person",
		"name": "Josiah Carberry",
		"affiliation": {
		  "@id": "https://ror.org/05gq02987"
		}
	  },
	  {
		"@id": "https://ror.org/05gq02987",
		"@type": "Organization",
		"name": "Brown University",
		"url": "http://www.brown.edu/"
	  }
	]
  }
`)

// specCrate is an example of a ro-crate-metadata.json file found in
// the wild.
var specCrate []byte = []byte(`
  { "@context": "https://w3id.org/ro/crate/1.1/context",
  "@graph": [

    {
      "@type": "CreativeWork",
      "@id": "ro-crate-metadata.json",
      "conformsTo": {"@id": "https://w3id.org/ro/crate/1.1"},
      "about": {"@id": "./"},
      "description": "RO-Crate Metadata File Descriptor (this file)"
    },
    {
      "@id": "./",
      "@type": "Dataset",
      "name": "Example RO-Crate",
      "description": "The RO-Crate Root Data Entity",
      "hasPart": [
        {"@id": "data1.txt"},
        {"@id": "data2.txt"}
      ]
    },
    {
      "@id": "data1.txt",
      "@type": "File",
      "description": "One of hopefully many Data Entities",
      "author": {"@id": "#alice"},
      "contentLocation":  {"@id": "http://sws.geonames.org/8152662/"}
    },
    {
      "@id": "data2.txt",
      "@type": "File"
    },

    {
      "@id": "#alice",
      "@type": "Person",
      "name": "Alice",
      "description": "One of hopefully many Contextual Entities"
    },
    {
      "@id": "http://sws.geonames.org/8152662/",
      "@type": "Place",
      "name": "Catalina Park"
    }
 ]
}
`)

// galaxyCrate is an example of a ro-crate-metadata.json file found
// in the wild.
var galaxyCrate []byte = []byte(`
{
    "@context": "https://w3id.org/ro/crate/1.1/context",
    "@graph": [
        {
            "@id": "ro-crate-metadata.json",
            "@type": "CreativeWork",
            "conformsTo": {
                "@id": "https://w3id.org/ro/crate/1.1"
            },
            "about": {
                "@id": "./"
            }
        },
        {
            "@id": "./",
            "@type": [
                "Dataset",
                "LearningResource"
            ],
            "name": "Demo Crate",
            "description": "a demo crate for Galaxy training",
            "datePublished": "2024-03-08",
            "publisher": "https://ror.org/0abcdef00",
            "license": {
                "@id": "https://spdx.org/licenses/CC-BY-NC-SA-4.0.html",
                "@type": "CreativeWork",
                "name": "CC BY-NC-SA 4.0 International",
                "description": "Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International"
            },
            "author": {
                "@id": "https://orcid.org/0000-0000-0000-0000"
            },
            "hasPart": [
                {
                    "@id": "data.csv"
                }
            ],
			"contentUrl": "http://example.com/resource/data"
        },
        {
            "@id": "data.csv",
            "@type": "File",
            "name": "Rainfall Katoomba 2022-02",
            "description": "Rainfall data for Katoomba in NSW Australia, captured February 2022.",
            "encodingFormat": "text/csv",
            "license": {
                "@id": "https://creativecommons.org/licenses/by-nc-sa/4.0/",
                "@type": "CreativeWork",
                "name": "CC BY-NC-SA 4.0 International",
                "description": "Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International"
            }
        },
        {
            "@id": "https://orcid.org/0000-0000-0000-0000",
            "@type": "Person",
            "givenName": "Jane",
            "familyName": "Smith",
            "affiliation": {
                "@id": "https://ror.org/0abcdef00"
            }
        },
        {
            "@id": "https://ror.org/0abcdef00",
            "@type": "Organization",
            "name": "Example University",
            "url": "https://www.example.org"
        }
    ]
}
`)

// emptyCrate provides a way to test the lacn of information, e.g.
// when providign summary information.
var emptyCrate []byte = []byte(`
{
    "@context": "https://w3id.org/ro/crate/1.1/context",
    "@graph": [
        {
            "@type": "CreativeWork",
            "@id": "ro-crate-metadata.json",
            "conformsTo": {
                "@id": "https://w3id.org/ro/crate/1.1"
            },
            "about": {
                "@id": "./"
            }
        },
        {
        },
        {
        }
    ]
}
`)
