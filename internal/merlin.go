package hopla

type MerlinSettings struct {
	RunMerlin         bool   //Run Merlin haplotyping
	Model             string //Merlin Haplotyping model [sample, best]
	MinSegVar         int    //Minimum number of variants in a same-haplotype segment
	MinSegVarX        int    //Minimum number of variants in a same-haplotype segment for X chromosome
	VotingWindowSize  int    //Size (in bp) to correct haplotypes by 'weighted neighbourhood voting
	VotingWindowSizeX int    //Size (in bp) to correct haplotypes by 'weighted neighbourhood voting for X chromosome
	ConcordanceTable  bool   //Create concordance table
}

func RunMerlin(settings *MerlinSettings, variants string) error {
	// Covert variants to merlin inputs

	// Run Merlin

	// Parse Merlin output

	return nil
}
