package hopla

type HoplaSettings struct {
	Family     string
	Samples    []string
	Fathers    []string
	Mothers    []string
	Sexes      []string
	Filter1    HoplaFilter1Settings
	Filter2    HoplaFilter2Settings
	Merlin     MerlinSettings
	Annotation HoplaAnnotationSettings
}

type HoplaFilter1Settings struct {
	//dp.hard.limit
	//Minimum depth of coverage for all samples, unless DepthHardLimitIds is specified.
	//If a variant does not meet this requirement, it will be filtered out.
	DepthHardLimit int
	//dp.hard.limit.ids
	//Samples to apply the hard limit to
	DepthHardLimitIds []string

	//dp.soft.limit
	//Minimum depth of coverage for all samples, unless DepthSoftLimitIds is specified.
	//If a variant does not meet this requirement, only the sample that does not comply will be filtered out.
	DepthSoftLimit int
	//dp.soft.limit.ids
	//Samples to apply the soft limit to
	DepthSoftLimitIds []string

	//af.hard.limit
	// Minimum allele frequency for all samples, unless AlleleFrequencyHardLimitIds is specified.
	// If a variant does not meet this requirement, it will be filtered out.
	AlleleFrequencyHardLimit float32
	//af.hard.limit.ids
	// Samples to apply the hard limit to
	AlleleFrequencyHardLimitIds []string
}

type HoplaFilter2Settings struct {
	//keep.informative.ids
	//keep.hetero.ids
}

type HoplaAnnotationSettings struct {
	regions       []string
	referenceIds  []string
	carrierIds    []string
	affectedIds   []string
	unaffectedIds []string
	info          []string
}

// Return a new HoplaSettings struct with sensible defaults
func (settings HoplaSettings) New() *HoplaSettings {
	return &HoplaSettings{
		Family: "Hopla",
		Filter1: HoplaFilter1Settings{
			DepthHardLimit:           10,
			DepthSoftLimit:           10,
			AlleleFrequencyHardLimit: 0.0,
		},
		Filter2: HoplaFilter2Settings{},
		Merlin: MerlinSettings{
			RunMerlin:        true,
			Model:            "best",
			MinSegVar:        5,
			MinSegVarX:       15,
			VotingWindowSize: 10000000,
			ConcordanceTable: true,
		},
		Annotation: HoplaAnnotationSettings{},
	}
}

func (settings *HoplaSettings) Load(settingsFile string) {
	// Load settings from yaml file
}
