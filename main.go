package main

import (
	"context"
	
	"fmt"
	

	"math/rand"
	"os"
//	"path/filepath"
	
	"bufio"
	"github.com/ipfs/go-cid"
	"strconv"

	saproof "github.com/filecoin-project/specs-actors/actors/runtime/proof"

	"github.com/docker/go-units"
	logging "github.com/ipfs/go-log/v2"
//	"github.com/minio/blake2b-simd"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"
	paramfetch "github.com/filecoin-project/go-paramfetch"
	"github.com/filecoin-project/go-state-types/abi"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper/basicfs"
//	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/actors/policy"
//	"github.com/filecoin-project/specs-storage/storage"

//	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
)

var log = logging.Logger("lotus-bench")

type Commit2In struct {
        SectorNum  int64
        Phase1Out  []byte
        SectorSize uint64
}


func main() {
	logging.SetLogLevel("*", "INFO")

	log.Info("Starting lotus-bench")

	// miner.SupportedProofTypes[abi.RegisteredSealProof_StackedDrg2KiBV1] = struct{}{}

	policy.AddSupportedProofTypes(abi.RegisteredSealProof_StackedDrg2KiBV1)

	app := &cli.App{
		Name:    "sector-check",
		Usage:   "check window post",
		Version: build.UserVersion(),
		Commands: []*cli.Command{
			sealBenchCmd,
			importBenchCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}

var sealBenchCmd = &cli.Command{
	Name: "checking",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "storage-dir",
			Value: "~/.lotus-bench",
			Usage: "Path to the storage directory that will store sectors long term",
		},
		&cli.StringFlag{
			Name:  "sector-size",
			Value: "512MiB",
			Usage: "size of the sectors in bytes, i.e. 32GiB",
		},
                &cli.StringFlag{
                        Name:  "sectors-file",
                        Value: "sectors.txt",
                        Usage: "absolute path file. contains line number, line cidcommr, line number...",
                },
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the checking",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t010010",
		},
                &cli.IntFlag{
                        Name:  "number",
                        Value: 1,
                },

                &cli.StringFlag{
                        Name:  "cidcommr",
			Usage: "CIDcommR,  eg/default.  bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z",
			Value: "bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z",
                },
	},
	Action: func(c *cli.Context) error {
		if c.Bool("no-gpu") {
			err := os.Setenv("BELLMAN_NO_GPU", "1")
			if err != nil {
				return xerrors.Errorf("setting no-gpu flag: %w", err)
			}
		}

		var sbdir string

		sdir, err := homedir.Expand(c.String("storage-dir"))
		if err != nil {
			return err
		}

		err = os.MkdirAll(sdir, 0775) //nolint:gosec
		if err != nil {
			return xerrors.Errorf("creating sectorbuilder dir: %w", err)
		}

		defer func() {
		}()

		sbdir = sdir

		// miner address
		maddr, err := address.NewFromString(c.String("miner-addr"))
		if err != nil {
			return err
		}
		log.Infof("miner maddr: ", maddr)
		amid, err := address.IDFromAddress(maddr)
		if err != nil {
			return err
		}
		log.Infof("miner amid: ", amid)
		mid := abi.ActorID(amid)
		log.Infof("miner mid: ", mid)

		// sector size
		sectorSizeInt, err := units.RAMInBytes(c.String("sector-size"))
		if err != nil {
			return err
		}
		sectorSize := abi.SectorSize(sectorSizeInt)

		// spt, err := ffiwrapper.SealProofTypeFromSectorSize(sectorSize)
		spt := spt(sectorSize)
		

		// cfg := &ffiwrapper.Config{
		// 	SealProofType: spt,
		// }

		if err := paramfetch.GetParams(lcli.ReqContext(c), build.ParametersJSON(), uint64(sectorSize)); err != nil {
			return xerrors.Errorf("getting params: %w", err)
		}

		sbfs := &basicfs.Provider{
			Root: sbdir,
		}

		// sb, err := ffiwrapper.New(sbfs, spt)
		sb, err := ffiwrapper.New(sbfs)
		if err != nil {
			return err
		}


		// sealedSectors := getSectorsInfo(c.String("sectors-file"), sb.SealProofType())
		sealedSectors := getSectorsInfo(c.String("sectors-file"), spt)

		var challenge [32]byte
		rand.Read(challenge[:])

		log.Info("computing window post snark (cold)")
		wproof1, ps, err := sb.GenerateWindowPoSt(context.TODO(), mid, sealedSectors, challenge[:])
		if err != nil {
			return err
		}

		wpvi1 := saproof.WindowPoStVerifyInfo{
			Randomness:        challenge[:],
			Proofs:            wproof1,
			ChallengedSectors: sealedSectors,
			Prover:            mid,
		}

		log.Info("generate window PoSt skipped sectors", "sectors", ps, "error", err)

		ok, err := ffiwrapper.ProofVerifier.VerifyWindowPoSt(context.TODO(), wpvi1)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("post verification failed")
		}

		return nil
	},
}

func getSectorsInfo(filePath string, proofType abi.RegisteredSealProof) []saproof.SectorInfo {

	sealedSectors := make([]saproof.SectorInfo, 0)

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return sealedSectors
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		sectorIndex := scanner.Text()

		index, error := strconv.Atoi(sectorIndex)
		if error != nil {
			fmt.Println("error")
			break
		}

		scanner.Scan()
		cidStr := scanner.Text()
		ccid, err := cid.Decode(cidStr)
                if(err != nil) {
                        log.Infof("cid error, ignore sectors after this: %d, %s", uint64(index), err)
			return sealedSectors 
                }

		var sector saproof.SectorInfo
		sector.SealProof = proofType
		sector.SectorNumber = abi.SectorNumber(uint64(index))
		sector.SealedCID = ccid

		sealedSectors = append(sealedSectors, sector)

		log.Infof("id: ", sector.SectorNumber)
		log.Infof("cid: ", sector.SealedCID)

	}

	fmt.Println("sector length", len(sealedSectors))
	return sealedSectors
}



func spt(ssize abi.SectorSize) abi.RegisteredSealProof {
	spt, err := miner.SealProofTypeFromSectorSize(ssize, build.NewestNetworkVersion)
	if err != nil {
		panic(err)
	}

	return spt
}

