package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	hopla "github.com/CenterForMedicalGeneticsGhent/Hopla/internal"
	"github.com/biogo/hts/bgzf"
	"github.com/brentp/irelate/parsers"
	"github.com/brentp/vcfgo"
	"github.com/brentp/xopen"
	"github.com/urfave/cli/v2"
)

const version = "2.0.0"

func main() {
	app := &cli.App{
		Name:      "hopla",
		Version:   version,
		Authors:   []*cli.Author{{Name: "CMGG Bioinformatics Team", Email: "ict.cmgg@uzgent.be"}},
		Copyright: "2024 Center for Medical Genetics Ghent, Ghent University Hospital",
		Usage:     "Haplotype analysis",
		UsageText: "hopla --settings <settings.yaml> <vcf>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "settings",
				Aliases:  []string{"s"},
				Usage:    "settings `YAML` file",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "cytoband",
				Aliases: []string{"c"},
				Usage:   "UCSC cytoband `TXT` file",
			},
		},
		Before: func(c *cli.Context) error {
			// Santiy checks
			// Check if settings file exists
			if _, err := os.Stat(c.String("settings")); os.IsNotExist(err) {
				log.Fatalf("Settings file %s does not exist", c.String("settings"))
			}
			// Check if VCF file exists
			if _, err := os.Stat(c.Args().First()); os.IsNotExist(err) {
				log.Fatalf("VCF file %s does not exist", c.Args().First())
			}
			// Check if cytoband file exists
			if c.String("cytoband") != "" {
				if _, err := os.Stat(c.String("cytoband")); os.IsNotExist(err) {
					log.Fatalf("Cytoband file %s does not exist", c.String("cytoband"))
				}
			}
			// Check if Merlin is installed
			// _, err := exec.LookPath("merlin")
			// if err != nil {
			// 	log.Fatalf("Merlin is not installed")
			// }
			return nil
		},
		Action: func(c *cli.Context) error {
			// Load settings yaml file
			settings := hopla.HoplaSettings{}.New()

			// Get VCF file
			vcf := c.Args().First()

			// Get cytoband file
			cytoband := c.String("cytoband")

			RunHopla(settings, vcf, cytoband)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func RunHopla(settings *hopla.HoplaSettings, vcf string, cytoband string) (err error) {
	// Load cytobands
	if cytoband != "" {
		_, err := hopla.ReadCytobands(cytoband)
		if err != nil {
			log.Println("Unable to parse Cytobands file: ", err)
		}
	}

	// Load VCF
	var vcfReader io.Reader
	// code adapted from vcfanno (https://github.com/brentp/vcfanno/blob/master/vcfanno.go#L130-L161)
	// try to parallelize reading if we have plenty of CPUs and it's (possibly)
	// a bgzf file.
	if strings.HasSuffix(vcf, ".gz") || strings.HasSuffix(vcf, ".bgz") {
		if rdr, err := os.Open(vcf); err == nil {
			if st, err := rdr.Stat(); err == nil && st.Size() > 2320303098 {
				vcfReader, err = bgzf.NewReader(rdr, 4)
				if err == nil {
					log.Printf("using 4 worker threads to decompress bgzip file")
				} else {
					vcfReader = nil
				}
			} else {
				vcfReader, err = bgzf.NewReader(rdr, 2)
				if err == nil {
					log.Printf("using 2 worker threads to decompress bgzip file")
				} else {
					vcfReader = nil
				}
			}
		} else {
			log.Fatal(err)
		}
	}
	if vcfReader == nil {
		vcfReader, err = xopen.Ropen(vcf)
		log.Printf("falling back to non-bgzip")
	}
	if err != nil {
		log.Fatal(fmt.Errorf("error opening vcf file %s: %s", vcf, err))
	}

	_, query, err := parsers.VCFIterator(vcfReader)
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing vcf file %s: %s", vcf, err))
	}

	// Get the sequence dictionary from the vcf to get a list of chromosomes
	sequenceDict := query.Header.Contigs
	log.Printf("%v contigs found in VCF ", len(sequenceDict))

	// Get the samples from the vcf
	vcfSamples := query.Header.SampleNames
	log.Println("Samples in VCF: ", vcfSamples)

	for {
		variant := query.Read()
		if variant == nil {
			break
		}

		// Check if variant is a SNP
		// Naive filter to check if the variant has a single alternate allele
		if len(variant.Reference) != 1 || len(variant.Alt()) != 1 {
			continue
		}

		// Apply Allele Frequency filter
		AF, err := variant.Info().Get("AF")
		if err != nil {
			log.Println("Variant does not have an AF field: ", variant)
		}
		// if the allele frequency is higher than the hard limit, the variant is filtered out
		if AF.([]float32)[0] >= settings.Filter1.AlleleFrequencyHardLimit {
			continue
		}

		// Apply Depth filter
		err = query.Header.ParseSamples(variant)
		if err != nil {
			log.Println("Error parsing samples: ", err)
		}
		sample := variant.Samples[0]
		DP, err := variant.GetGenotypeField(sample, "DP", 0)
		if err != nil {
			log.Println("Variant does not have a DP field: ", variant)
		}
		fmt.Println(DP)
	}
	// Check if all samples have a sex assigned
	// If any of the samples has an "NA" sex, infer it from the variants
	err = PredictSex()
	if err != nil {
		return err
	}

	// Run Merlin
	err = hopla.RunMerlin(&settings.Merlin, vcf)
	if err != nil {
		return err
	}

	return nil
}

func DepthFilter(settings *hopla.HoplaFilter1Settings, variant *vcfgo.Variant) bool {
	err := variant.Header.ParseSamples(variant)
	if err != nil {
		log.Println("Error parsing samples: ", err)
	}
	fmt.Println(variant.Samples[0])
	//empty return

	// for i := range variant.Header.ParseSamples(variant) {
	// 	v := variant.Samples[i]
	// 	fmt.Println(v)
	// 	DP, err := variant.GetGenotypeField(v, "DP", 0)
	// 	if err != nil {
	// 		log.Println("Variant does not have a DP field: ", variant)
	// 	}
	// 	fmt.Println(DP)
	// }
	return true
}

func HoplaFilter2(settings *hopla.HoplaFilter2Settings, variant *vcfgo.Variant) bool {
	return true
}

func PredictSex() error {
	return nil
}
