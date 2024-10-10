package hopla

type MerlinModel string

const (
	MerlinSampleModel MerlinModel = "sample"
	MerlinBestModel   MerlinModel = "best"
)

// MerlinSettings contains the settings for Merlin haplotyping
type MerlinSettings struct {
	Model             MerlinModel `yaml:"model"`                //Merlin Haplotyping model [sample, best]
	MinSegVar         int         `yaml:"min.seg.var"`          //Minimum number of variants in a same-haplotype segment
	MinSegVarX        int         `yaml:"min.seg.var.X"`        //Minimum number of variants in a same-haplotype segment for X chromosome
	VotingWindowSize  int         `yaml:"window.size.voting"`   //Size (in bp) to correct haplotypes by weighted neighbourhood voting
	VotingWindowSizeX int         `yaml:"window.size.voting.X"` //Size (in bp) to correct haplotypes by weighted neighbourhood voting for X chromosome
	KeepChromosomes   bool        `yaml:"keep.chromosomes.only"`
	KeepRegions       bool        `yaml:"keep.regions.only"`
}

func RunMerlin(settings *MerlinSettings, variants string) error {
	// Covert variants to merlin inputs

	// Run Merlin

	// Parse Merlin output

	return nil
}
