package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

type (
	scannerOptions map[string]string

	cliFlags struct {
		command    string
		device     string
		source     string
		options    scannerOptions
		verbose    bool
		fileFormat string //file fileFormat: png, jpg, pdf
	}
)

const (
	optionFilterOnly = "FILTER_ONLY"
	cmdPrintScanners = "print-scanners"
	cmdPrintOptions  = "print-options"
	cmdScan          = "scan"
	generalUsage     = `usage: %s #{cmdPrintScanners}|#{cmdPrintOptions}|#{cmdScan}

Commands:
#{cmdPrintScanners}: find and print available scanners
#{cmdPrintOptions}: print scanner and paper source options
#{cmdScan}: scan using specified scanner and paper source
`
)

var (
	cmdUsage = map[string]string{
		cmdPrintScanners: `usage: %s #{cmdPrintScanners} [-v]
Find and print available scanners`,
		cmdPrintOptions: `usage: %s #{cmdPrintOptions} [-d scanner] [-s paper_source] [filter options] [-v]
Print scanner and paper source options

Options:
`,
		cmdScan: `usage: %s #{cmdScan} [-d scanner] [-s paper_source] [scan options] [-f file format] [-v]
Scan using specified scanner and paper source. Output file will have name like 'page1.png, page2.jpg or result.pdf' depending on -f option value. 

Options:
`,
	}

	r = strings.NewReplacer("#{cmdPrintScanners}", cmdPrintScanners, "#{cmdPrintOptions}", cmdPrintOptions, "#{cmdScan}", cmdScan)
)

func (f *scannerOptions) String() string {
	return "scannerOptions"
}

func (f *scannerOptions) Set(value string) error {
	if value == "" {
		panic("Invalid option value!")
	}
	r := strings.Split(value, "=")
	if len(r) > 2 {
		panic("Invalid option value!")
	}

	if len(r) == 1 {
		// 'key' form, do not set value, filter only
		if len(r[0]) == len(value) {
			(*f)[r[0]] = optionFilterOnly
			return nil
		}
		// 'key=' form, set value to empty string
		(*f)[r[0]] = ""
		return nil
	}

	(*f)[r[0]] = r[1]
	return nil
}

func addCommonFlags(fs *flag.FlagSet, flags *cliFlags) {
	fs.BoolVar(&flags.verbose, "v", false, "show debug messages")
}

func parseFlags() *cliFlags {
	exec := filepath.Base(os.Args[0])

	//general usage print
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), r.Replace(generalUsage), exec)
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) < 2 {
		flag.Usage()
		log.Fatalf("error: wrong number of arguments")
	}

	var flags cliFlags
	flags.command = os.Args[1]
	flags.options = scannerOptions{}
	var fs *flag.FlagSet

	switch flags.command {
	case cmdScan:
		fs = flag.NewFlagSet(cmdScan, flag.ExitOnError)
		addCommonFlags(fs, &flags)
		fs.StringVar(&flags.device, "d", "", "id of the scanner, mandatory")
		fs.StringVar(&flags.source, "s", "", "paper source, mandatory")
		fs.Var(&flags.options, "o", `try to set specified option before scan. 
Format:
-o name=value :  set option with [name] to [value]
-o name= : pass empty string as value of the option
This flag can appear multiple times: -o name1=value1 -o name2=value2`)
		fs.StringVar(&flags.fileFormat, "f", "pdf", "output file format [jpg|png|pdf]")
		fs.Usage = func() {
			fmt.Fprintf(flag.CommandLine.Output(), r.Replace(cmdUsage[cmdScan]), exec)
			fs.PrintDefaults()
		}

		if err := fs.Parse(os.Args[2:]); err != nil {
			fs.Usage()
			log.Fatalf(err.Error())
		}

		if flags.device == "" || flags.source == "" {
			fs.Usage()
			log.Fatalf(r.Replace("#{cmdScan}: invalid command"))
		}
		switch flags.fileFormat {
		case "png","jpg","pdf":
		default:
			fs.Usage()
			log.Fatalf(r.Replace("#{cmdScan}: invalid file format"))
		}

		return &flags

	case cmdPrintOptions:

		fs = flag.NewFlagSet(cmdPrintOptions, flag.ExitOnError)
		addCommonFlags(fs, &flags)
		fs.StringVar(&flags.device, "d", "", "id of the scanner, mandatory")
		fs.StringVar(&flags.source, "s", "", "paper source, mandatory")
		fs.Var(&flags.options, "o", `try to set specified option before printing.
If this flag is specified, the output will be filtered by provided options.

Format:
-o name=value :  set option with [name] to [value]
-o name= : pass empty string as value of the option
-o name : just filter options list by the option name
This flag can appear multiple times: -o name1=value1 -o name2=value2`)
		fs.Usage = func() {
			fmt.Fprintf(flag.CommandLine.Output(), r.Replace(cmdUsage[cmdPrintOptions]), exec)
			fs.PrintDefaults()
		}

		if err := fs.Parse(os.Args[2:]); err != nil {
			fs.Usage()
			log.Fatalf(err.Error())
		}

		if flags.device == "" || flags.source == "" {
			fs.Usage()
			log.Fatalf(r.Replace("#{cmdPrintOptions}: invalid command"))
		}

		return &flags

	case cmdPrintScanners:
		fs = flag.NewFlagSet(cmdPrintScanners, flag.ExitOnError)
		addCommonFlags(fs, &flags)
		fs.Usage = func() {
			fmt.Fprintf(flag.CommandLine.Output(), r.Replace(cmdUsage[cmdPrintScanners]), exec)
			fs.PrintDefaults()
		}
		if len(os.Args) > 2 {
			fs.Usage()
			log.Fatalf("scan: wrong number of arguments\n\n")
		}
		if err := fs.Parse(os.Args[2:]); err != nil {
			fs.Usage()
			log.Fatalf(err.Error())
		}
		return &flags

	default:
		flag.Usage()
		log.Fatalf("error: incorrect command")
	}

	return nil
}
