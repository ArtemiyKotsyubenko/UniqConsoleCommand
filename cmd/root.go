package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	tooManyFlagsErrMsg = "Flags -c -u -d cannot be used at the same time"
	count              = "count"
	repeated           = "repeated"
	unique             = "unique"
	ignoreCase         = "ignore-case"
	skipFields         = "skip-fields"
	skipChars          = "skip-chars"
)

var (
	inputFile    *os.File  = nil
	outputFile   *os.File  = nil
	inputStream  io.Reader = os.Stdin
	outputStream io.Writer = os.Stdout
	cFlag        int8      = 0
	dFlag        int8      = 0
	uFlag        int8      = 0
	iFlag        int8      = 0
	fFlagVal     uint64    = 0
	sFlagVal     uint64    = 0
)

func Ignore(...interface{}) {
	//do nothing
}

func LogIfError(args ...interface{}) {
	for _, arg := range args {
		switch arg.(type) {
		case error:
			Ignore(fmt.Fprintln(os.Stderr, arg))
		}
	}
}

func TruncateNFields(str string, n uint64) string {
	re := regexp.MustCompile(`\s*\S*`)
	for n != 0 && len(str) != 0 {
		str = str[len(re.FindString(str)):]
		n--
	}
	return str
}

func TruncateNCharacters(str string, n uint64) string {
	ans := ""
	if n <= uint64(len(str)) {
		ans = str[n:]
	}
	return ans
}

func GetStringToCompare(str string) string {
	str = TruncateNCharacters(TruncateNFields(str, fFlagVal), sFlagVal)
	if iFlag != 0 {
		str = strings.ToLower(str)
	}
	return str
}

type StreamPrinter interface {
	AddString(string)
	FinishWork()
}

type DefaultPrinter struct {
	firstStringFromInput bool // = true by default
	prevFormattedStr     string
}

func (p *DefaultPrinter) AddString(currStr string) {
	currFormattedStr := GetStringToCompare(currStr)
	if currFormattedStr != p.prevFormattedStr || p.firstStringFromInput {
		p.firstStringFromInput = false
		p.prevFormattedStr = currFormattedStr
		LogIfError(fmt.Fprintln(outputStream, currStr))
	}
}

func (p *DefaultPrinter) FinishWork() {
	// no need to do anything
}

type CountingPrinter struct {
	firstStringFromInput bool
	stringToPrint        string
	prevFormattedStr     string
	counter              int64 // = -1 by default
}

func (p *CountingPrinter) AddString(currStr string) {
	currFormattedStr := GetStringToCompare(currStr)

	p.counter++
	if !p.firstStringFromInput && currFormattedStr != p.prevFormattedStr {
		LogIfError(fmt.Fprintf(outputStream, "      %d %s\n", p.counter, p.stringToPrint))
		p.counter = 0
	}
	if p.counter == 0 {
		p.stringToPrint = currStr
	}

	p.prevFormattedStr = currFormattedStr
	p.firstStringFromInput = false
}

func (p *CountingPrinter) FinishWork() {
	p.counter++
	LogIfError(fmt.Fprintf(outputStream, "      %d %s\n", p.counter, p.stringToPrint))
}

type PrinterForRepeated struct {
	firstStringFromInput bool
	stringToPrint        string
	prevFormattedStr     string
	counter              int64
}

func (p *PrinterForRepeated) AddString(currStr string) {
	if p.counter == 0 {
		p.stringToPrint = currStr
	}
	currFormattedString := GetStringToCompare(currStr)
	if currFormattedString == p.prevFormattedStr || p.firstStringFromInput {
		p.counter++
	} else {
		if p.counter != 0 {
			LogIfError(fmt.Fprintln(outputStream, p.stringToPrint))
		}
		p.counter = 0
	}
	p.prevFormattedStr = currFormattedString
	p.firstStringFromInput = false
}

func (p *PrinterForRepeated) FinishWork() {
	if p.counter != 0 {
		LogIfError(fmt.Fprintln(outputStream, p.stringToPrint))
	}
}

type PrinterForUniq struct {
	firstStringFromInput bool
	stringToPrint        string
	prevFormattedStr     string
	counter              int64 // = -1 by default
}

func (p *PrinterForUniq) AddString(currStr string) {
	currFormattedStr := GetStringToCompare(currStr)
	p.counter++
	if !p.firstStringFromInput && currFormattedStr != p.prevFormattedStr {
		if p.counter == 1 {
			LogIfError(fmt.Fprintln(outputStream, p.stringToPrint))
		}
		p.counter = 0
	}
	if p.counter == 0 || p.firstStringFromInput {
		p.stringToPrint = currStr
	}
	p.prevFormattedStr = currFormattedStr
	p.firstStringFromInput = false
}

func (p *PrinterForUniq) FinishWork() {
	p.counter++
	if p.counter == 1 {
		LogIfError(fmt.Fprintln(outputStream, p.stringToPrint))
	}
}

func GetPrinter() StreamPrinter {
	if cFlag != 0 {
		return &CountingPrinter{
			firstStringFromInput: true,
			stringToPrint:        "",
			prevFormattedStr:     "",
			counter:              -1,
		}
	} else if dFlag != 0 {
		return &PrinterForRepeated{
			firstStringFromInput: true,
			stringToPrint:        "",
			prevFormattedStr:     "",
			counter:              0,
		}
	} else if uFlag != 0 {
		return &PrinterForUniq{
			firstStringFromInput: true,
			stringToPrint:        "",
			prevFormattedStr:     "",
			counter:              -1,
		}
	}
	return &DefaultPrinter{
		firstStringFromInput: true,
		prevFormattedStr:     "",
	}
}

func CloseFiles() {
	if inputFile != nil {
		LogIfError(inputFile.Close())
	}
	if outputFile != nil {
		LogIfError(outputFile.Close())
	}
}

func Prepare(cmd *cobra.Command, args []string) {

	ReadFlag := func(name string) int8 {
		flag, _ := cmd.Flags().GetBool(name)
		ans := int8(0)
		if flag {
			ans = 1
		}
		return ans
	}

	var err error

	cFlag = ReadFlag(count)
	dFlag = ReadFlag(repeated)
	uFlag = ReadFlag(unique)
	iFlag = ReadFlag(ignoreCase)
	fFlagVal, err = cmd.Flags().GetUint64(skipFields)
	sFlagVal, err = cmd.Flags().GetUint64(skipChars)

	if cFlag+dFlag+uFlag > 1 {
		Ignore(fmt.Fprintln(os.Stderr, tooManyFlagsErrMsg))
		os.Exit(1)
	}

	if len(args) >= 1 {
		inputFile, err = os.Open(args[0])
		inputStream = inputFile

		if err != nil {
			Ignore(fmt.Fprintln(os.Stderr, err))
			os.Exit(1)
		}
	}
	if len(args) == 2 {
		outputFile, err = os.OpenFile(args[1], os.O_CREATE|os.O_WRONLY, os.ModePerm)
		LogIfError(outputFile.Truncate(0))
		outputStream = outputFile
		if err != nil {
			Ignore(fmt.Fprintln(os.Stderr, err))
			os.Exit(1)
		}
	}
}

var rootCmd = &cobra.Command{
	Use:     "uniq [-c | -d | -u] [-i] [-f num] [-s chars] [input_file [output_file]]",
	Short:   "UNIX uniq analog",
	Long:    "Filter adjacent matching lines from INPUT (or standard input),\nwriting to OUTPUT (or standard output).",
	Version: "1.0",
	Args:    cobra.MaximumNArgs(2),

	Run: func(cmd *cobra.Command, args []string) {
		defer CloseFiles()
		Prepare(cmd, args)

		scanner := bufio.NewScanner(inputStream)
		printer := GetPrinter()
		for scanner.Scan() {
			currStr := scanner.Text()
			printer.AddString(currStr)
		}
		printer.FinishWork()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		Ignore(fmt.Fprintln(os.Stderr, err))
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP(count, "c", false, "prefix lines by the number of occurrences")
	rootCmd.Flags().BoolP(repeated, "d", false, "only print duplicate lines, one for each group")
	rootCmd.Flags().BoolP(unique, "u", false, "only print unique lines")
	rootCmd.Flags().BoolP(ignoreCase, "i", false, "ignore differences in case when comparing")
	rootCmd.Flags().Uint64P(skipFields, "f", 0, "avoid comparing the first N fields")
	rootCmd.Flags().Uint64P(skipChars, "s", 0, "avoid comparing the first N characters")
}
