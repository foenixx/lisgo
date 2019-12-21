package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"image/jpeg"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/foenixx/lisgo"
	"github.com/oliverpool/gofpdf"
)

func printOption(o *lisgo.OptionDescriptor) {
	color.Cyan("------- %s ------", o.Name)
	fmt.Println(o)
}

func printOptions(device string, source string, options *scannerOptions) {
	log.WithField("scanner", device).WithField("paper source", source).Info("printing options")

	lis, err := lisgo.New()
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	d, err := lis.GetDevice(device)
	if err != nil {
		panic(err)
	}

	err = d.Open()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	s, err := d.GetPaperSource(source)
	if err != nil {
		panic(err)
	}
	if s == nil {
		log.WithField("paper source", source).Error("cannot find paper source")
		d.Close()
		return
	}

	for key, val := range *options {
		if val == optionFilterOnly {
			continue
		}
		err = s.SetOption(key, val)
	}

	if len(*options) == 0 {
		err = s.IterateOptions(func(o *lisgo.OptionDescriptor) bool {
			printOption(o)
			return true
		})

		if err != nil {
			panic(err)
		}
		return
	}

	err = s.IterateOptions(func(o *lisgo.OptionDescriptor) bool {
		if _, ok := (*options)[o.Name]; ok {
			printOption(o)
		}
		return true
	})
	if err != nil {
		panic(err)
	}

}

func printScanners() {
	lis, err := lisgo.New()
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	scanners, err := lis.ListDevices()
	if err != nil {
		panic(err)
	}

	for _, d := range scanners {
		color.Cyan("---------------------------------------------------------------------\n")
		fmt.Printf("Device Id: ")
		color.Cyan(d.DeviceID)
		fmt.Printf("\nVendor: %v\nModel: %v\n", d.Vendor, d.Model)

		err = d.Open()
		if err != nil {
			log.WithError(err).Error("cannot acquire scanner")
			continue
		}

		err = d.IterateSources(func(s *lisgo.PaperSource) bool {
			fmt.Printf("Paper source: ")
			color.Green(s.Name)
			return true
		})
		if err != nil {
			log.WithError(err).Error("cannot get paper sources")
		}
		d.Close()
	}
}

func scanToImage(device string, source string, fileFormat string, options *scannerOptions) {
	lis, err := lisgo.New()
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	scanner, err := lis.GetDevice(device)

	if err != nil {
		panic(err)
	}
	err = scanner.Open()
	if err != nil {
		panic(err)
	}
	defer scanner.Close()

	ps, err := scanner.GetPaperSource(source)
	if err != nil {
		panic(err)
	}

	for key, val := range *options {
		if val == optionFilterOnly {
			continue
		}
		err = ps.SetOption(key, val)
	}

	var page *lisgo.PageReader

	var params *lisgo.ScanParameters
	session, err := ps.ScanStart()
	if err != nil {
		panic(err)
	}
	pageNum := 1

	for !session.EndOfFeed() {

		params, err = session.GetScanParameters()
		if err != nil {
			panic(err)
		}
		page = lisgo.NewPageReader(session, params)
		log.WithFields(log.Fields{
			"fileFormat": params.ImageFormatStr(),
			"w": params.Width(),
			"h": params.Height(),
			"image_size": params.ImageSize(),
		}).Debug("scanning parameters")


		imgName := fmt.Sprintf("page%d.%s", pageNum, fileFormat)
		err = page.WriteToFile(imgName, fileFormat)

		if err != nil {
			log.WithError(err).Error("cannot write output file")
			panic(err)
		}

		pageNum++
	}

}

func scanToPdf(device string, source string, options *scannerOptions) {
	lis, err := lisgo.New()
	if err != nil {
		panic(err)
	}
	defer lis.Close()
	scanner, err := lis.GetDevice(device)

	if err != nil {
		panic(err)
	}
	err = scanner.Open()
	if err != nil {
		panic(err)
	}
	defer scanner.Close()

	ps, err := scanner.GetPaperSource(source)
	if err != nil {
		panic(err)
	}

	for key, val := range *options {
		if val == optionFilterOnly {
			continue
		}
		err = ps.SetOption(key, val)
	}

	var page *lisgo.PageReader

	var params *lisgo.ScanParameters
	session, err := ps.ScanStart()
	if err != nil {
		panic(err)
	}
	pageNum := 1

	var opt = gofpdf.ImageOptions{ImageType: "jpeg"}

	pdf := gofpdf.New("P", "mm", "A4", "")

	for !session.EndOfFeed() {

		params, err = session.GetScanParameters()
		if err != nil {
			panic(err)
		}
		page = lisgo.NewPageReader(session, params)
		log.Debug("---- parameters -----")
		log.Debugf("%+v", params)

		pdf.AddPage()

		var buf = bytes.Buffer{}
		img, err := page.GetImage()
		if err != nil {
			panic(err)
		}

		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 50})
		if err != nil {
			panic(err)
		}

		imgName := fmt.Sprintf("page%d.jpg", pageNum)
		pdf.RegisterImageOptionsReader(imgName, opt, &buf)
		//// A4 210.0 x 297.0
		pdf.ImageOptions(imgName, 0, 0, 210, 297, false, opt, 0, "")
		//err = page.WriteToJpeg(fmt.Sprintf("page%d.jpg", pageNum))
		pageNum++
	}
	err = pdf.OutputFileAndClose("result.pdf")
	if err != nil {
		log.WithError(err).Error("cannot write pdf")
		panic(err)
	}
}

func main() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.ErrorLevel)

	var flags *cliFlags

	flags = parseFlags()
	if flags.verbose {
		log.SetLevel(log.DebugLevel)
	}

	switch flags.command {
	case cmdPrintScanners:
		printScanners()
	case cmdPrintOptions:
		printOptions(flags.device, flags.source, &flags.options)
	case cmdScan:
		if flags.fileFormat == "pdf" {
			scanToPdf(flags.device, flags.source, &flags.options)
		} else {
			scanToImage(flags.device, flags.source, flags.fileFormat, &flags.options)
		}
	}
}
