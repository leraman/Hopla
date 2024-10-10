package hopla

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HoplaSettings struct {
	Family     string                  `yaml:"family,omitempty"`
	Samples    []string                `yaml:"samples,omitempty"`
	Fathers    []string                `yaml:"fathers,omitempty"`
	Mothers    []string                `yaml:"mothers,omitempty"`
	Sexes      []string                `yaml:"sexes,omitempty"`
	Filter1    HoplaFilter1Settings    `yaml:"filter1,omitempty"`
	Filter2    HoplaFilter2Settings    `yaml:"filter2,omitempty"`
	Merlin     MerlinSettings          `yaml:"merlin,omitempty"`
	Annotation HoplaAnnotationSettings `yaml:"annotation,omitempty"`

	knownSexSamples  []string
	unkownSexSamples []string
}

type HoplaFilter1Settings struct {
	//dp.hard.limit
	//Minimum depth of coverage for all samples, unless DepthHardLimitIds is specified.
	//If a variant does not meet this requirement, it will be filtered out.
	DepthHardLimit int `yaml:"dp.hard.limit,omitempty"`
	//dp.hard.limit.ids
	//Samples to apply the hard limit to
	DepthHardLimitIds []string `yaml:"dp.hard.limit.ids,omitempty"`

	//dp.soft.limit
	//Minimum depth of coverage for all samples, unless DepthSoftLimitIds is specified.
	//If a variant does not meet this requirement, only the sample that does not comply will be filtered out.
	DepthSoftLimit int `yaml:"dp.soft.limit,omitempty"`
	//dp.soft.limit.ids
	//Samples to apply the soft limit to
	DepthSoftLimitIds []string `yaml:"dp.soft.limit.ids,omitempty"`

	//af.hard.limit
	// Minimum allele frequency for all samples, unless AlleleFrequencyHardLimitIds is specified.
	// If a variant does not meet this requirement, it will be filtered out.
	AlleleFrequencyHardLimit float32 `yaml:"af.hard.limit,omitempty"`
	//af.hard.limit.ids
	// Samples to apply the hard limit to
	AlleleFrequencyHardLimitIds []string `yaml:"af.hard.limit.ids,omitempty"`
}

type HoplaFilter2Settings struct {
	//keep.informative.ids
	KeepInformativeIds []string `yaml:"keep.informative.ids,omitempty"`
	//keep.hetero.ids
	KeepHeteroIds []string `yaml:"keep.hetero.ids,omitempty"`
}

type HoplaAnnotationSettings struct {
	regions       []string `yaml:"regions,omitempty"`
	referenceIds  []string `yaml:"reference.ids,omitempty"`
	carrierIds    []string `yaml:"carrier.ids,omitempty"`
	affectedIds   []string `yaml:"affected.ids,omitempty"`
	unaffectedIds []string `yaml:"unaffected.ids,omitempty"`
	info          []string `yaml:"info,omitempty"`
}

type HoplaVisualisationSettings struct {
	BafIds []string `yaml:"baf.ids,omitempty"`
}

// Return a new HoplaSettings struct with sensible defaults
func (settings HoplaSettings) New() *HoplaSettings {
	return &HoplaSettings{
		Family: "Hopla",
		Filter1: HoplaFilter1Settings{
			DepthHardLimit:           10,
			DepthSoftLimit:           10,
			AlleleFrequencyHardLimit: 0.25,
		},
		Filter2: HoplaFilter2Settings{},
		Merlin: MerlinSettings{
			Model:            "best",
			MinSegVar:        5,
			MinSegVarX:       15,
			VotingWindowSize: 10000000,
		},
		Annotation: HoplaAnnotationSettings{},
	}
}

// Superimpose the settings from the settings file onto the default settings
func (settings *HoplaSettings) Load(settingsfile string) (err error) {
	// Load settings from file
	yamlFile, err := os.Open(settingsfile)
	if err != nil {
		return fmt.Errorf("Unable to open settings file: %s", err)
	}
	defer yamlFile.Close()

	yamlParser := yaml.NewDecoder(yamlFile)
	err = yamlParser.Decode(settings)
	if err != nil {
		return fmt.Errorf("Unable to parse settings file: %s", err)
	}
	return nil
}
