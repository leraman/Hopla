package hopla

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Cytoband struct {
	chromosome string
	start      int
	end        int
	band       string
	stain      string
}

func (cytoband Cytoband) New() *Cytoband {
	return &Cytoband{}
}

func ReadCytobands(cytobandsIdeo string) (cytobands []Cytoband, err error) {
	// Read cytoband file
	cytobandFile, err := os.Open(cytobandsIdeo)
	reader := csv.NewReader(cytobandFile)
	reader.Comma = '\t'
	defer cytobandFile.Close()

	// Parse cytoband file
	for {
		line, err := reader.Read()
		if err != nil {
			break
		}
		band := Cytoband{}.New()
		band.chromosome = line[0]
		band.start, err = strconv.Atoi(line[1])
		if err != nil {
			log.Error("unable to parse start position")
			break
		}
		band.end, err = strconv.Atoi(line[2])
		if err != nil {
			log.Error("unable to parse end position")
			break
		}
		band.band = line[3]
		band.stain = line[4]
		cytobands = append(cytobands, *band)
	}

	if len(cytobands) == 0 {
		return nil, fmt.Errorf("No valid cytobands found in file")
	}

	return cytobands, nil

}
