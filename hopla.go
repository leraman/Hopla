package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	hopla "github.com/CenterForMedicalGeneticsGhent/Hopla/internal"
	"github.com/biogo/hts/bgzf"
	"github.com/brentp/irelate/parsers"
	"github.com/brentp/vcfgo"
	"github.com/brentp/xopen"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const version = "2.0.0"

// TYPES
type PedSample struct {
	Family string
	Sample string
	Father string
	Mother string
	Sex    int
	Status int
}

type Ped []PedSample

// FUNCTIONS
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
				Name:        "loglevel",
				Usage:       "log level (debug, info, warn, error, fatal, panic)",
				Value:       "info",
				DefaultText: "info",
			},
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
			// Set log level
			switch c.String("loglevel") {
			case "debug":
				log.Info("Setting log level to debug")
				log.SetLevel(log.DebugLevel)
			case "info":
				log.Info("Setting log level to info")
				log.SetLevel(log.InfoLevel)
			case "warn":
				log.Info("Setting log level to warn")
				log.SetLevel(log.WarnLevel)
			case "error":
				log.Info("Setting log level to error")
				log.SetLevel(log.ErrorLevel)
			case "fatal":
				log.Info("Setting log level to fatal")
				log.SetLevel(log.FatalLevel)
			case "panic":
				log.Info("Setting log level to panic")
				log.SetLevel(log.PanicLevel)
			default:
				log.Error("Invalid log level, defaulting to info")
				log.SetLevel(log.InfoLevel)
			}
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
			// Check if Minx is installed
			// _, err = exec.LookPath("minx")
			// if err != nil {
			// 	log.Fatalf("Minx is not installed")
			// }
			return nil
		},
		Action: func(c *cli.Context) (err error) {
			// Create a new settings instance
			settings := hopla.HoplaSettings{}.New()

			// Load settings yaml file and superimpose on default settings
			err = settings.Load(c.String("settings"))
			if err != nil {
				log.Fatal(err)
			}

			// Get cytoband file
			if c.String("cytoband") != "" {
				cytoband := c.String("cytoband")
				// Load cytobands
				_, err := hopla.ReadCytobands(cytoband)
				if err != nil {
					log.Warnf("Cytobands will not be displayed: %s", err)
				}
			}

			// Get VCF file
			vcf := c.Args().First()
			// Load variants
			_, err = LoadVariants(settings, vcf)
			if err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func LoadVariants(settings *hopla.HoplaSettings, vcf string) (variants []*vcfgo.Variant, err error) {
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
					log.Debug("using 4 worker threads to decompress bgzip file")
				} else {
					vcfReader = nil
				}
			} else {
				vcfReader, err = bgzf.NewReader(rdr, 2)
				if err == nil {
					log.Debug("using 2 worker threads to decompress bgzip file")
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
		log.Debug("falling back to non-bgzip")
	}
	if err != nil {
		log.Fatalf("error opening vcf file %s: %s", vcf, err)
	}

	_, query, err := parsers.VCFIterator(vcfReader)
	if err != nil {
		log.Fatal(fmt.Errorf("error parsing vcf file %s: %s", vcf, err))
	}

	// Get the sequence dictionary from the vcf to get a list of chromosomes
	sequenceDict := query.Header.Contigs
	log.Infof("%v contigs found in VCF ", len(sequenceDict))

	// Get the samples from the vcf
	vcfSamples := query.Header.SampleNames
	log.Info("Samples in VCF: ", vcfSamples)

	// Check if samples in settings are in the VCF
	unknownSampleRegex := regexp.MustCompile(`^U\d+$`)
	for _, sample := range settings.Samples {
		if !slices.Contains(vcfSamples, sample) {
			// Check if the samples is an "unknown" sample
			// if the sample is not an "unknown" sample, return an error
			if !unknownSampleRegex.MatchString(sample) {
				return nil, fmt.Errorf("Sample %s not found in VCF", sample)

			}
		}
	}

	for {
		variant := query.Read()
		// if the variant is nil, we have reached the end of the file
		if variant == nil {
			break
		}

		// Check if variant is a SNP
		// Naive filter to check if the variant has a single alternate allele
		if len(variant.Reference) != 1 || len(variant.Alt()) != 1 {
			continue
		}

		// Parse sampleGenotypes from variant
		err = query.Header.ParseSamples(variant)
		if err != nil {
			log.Debugf("Error parsing samples for variant %s with err %s ", variant, err)
		}

		// Apply hard AF filter
		AF, err := variant.Info().Get("AF")
		if err != nil {
			log.Debug("Variant does not have an AF field: ", variant)
		}
		// if the allele frequency is higher or equal to the hard limit, the variant is filtered out
		if AF.([]float32)[0] < settings.Filter1.AlleleFrequencyHardLimit {
			continue
		}

		// Apply hard DP filter (Variant depth)
		DP, err := variant.Info().Get("DP")
		if err != nil {
			log.Debug("Variant does not have a DP field: ", variant)
		}
		// if the depth is lower than the hard limit, the variant is filtered out
		if DP.(int) < settings.Filter1.DepthHardLimit {
			continue
		}

		// Apply soft DP filter (GT depth)
		for i, sample := range variant.Samples {
			if !slices.Contains(settings.Filter1.DepthSoftLimitIds, vcfSamples[i]) {
				// if the sample is not in the list of samples to apply the soft limit to, skip the sample
				continue
			}
			GTDP, err := variant.GetGenotypeField(sample, "DP", 0)
			if err != nil {
				log.Debug("Variant does not have a DP field: ", variant)
				continue
			}
			// if the depth is lower than the soft limit, the sample is filtered out
			if GTDP.(int) < settings.Filter1.DepthSoftLimit {
				// remove sample from variant
				variant.Samples[i] = nil
			}
		}

		// If all samples are filtered out, skip the variant
		for _, sample := range variant.Samples {
			if sample != nil {
				break
			}
		}

		// Add variant to list of variants
		variants = append(variants, variant)
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("No valid variants found in VCF")
	}
	return variants, nil
}
