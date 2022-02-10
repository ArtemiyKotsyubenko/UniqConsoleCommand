package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
)

var (
	inputStream  io.Reader = os.Stdin
	outputStream io.Writer = os.Stdout
	counter      uint64    = 0
)

func PrintWithCountFlag(prevStr, currStr *string, firstString, lastString bool) {
	if !firstString {
		counter++
	}

	if !firstString && *currStr != *prevStr {
		fmt.Fprintf(outputStream, "\t%d %s\r\n", counter, *prevStr)
		counter = 0
	}
	if lastString {
		fmt.Fprintf(outputStream, "\t%d %s\r\n", counter, *prevStr)
	}
}

func PrintWithRepeatedFlag(prevStr, currStr *string, lastString bool) {
	counter++
	if counter != 1 && *prevStr != *currStr || lastString && counter != 1 {
		fmt.Fprintln(outputStream, *prevStr)
		counter = 0
	}
}

func PrintWithUniqueFlag(prevStr, currStr *string) {
	counter++
	if counter == 1 && *prevStr != *currStr {
		fmt.Fprintln(outputStream, *currStr)
		counter = 0
	}

}

func PrintWithNoFlag(prevStr, currStr *string) {
	if *prevStr != *currStr {
		fmt.Fprintln(outputStream, *currStr)
	}

}

func TruncateNFields(str string, n uint64) string {
	idx := 0
	for l := len(str); idx < l; idx++ {
		if str[idx] == ' ' || str[idx] == '\t' {
			n--
		}
	}
	ans := ""
	if n == 0 {
		ans = str[idx:]
	}
	return ans
}

func TruncateNCharacters(str string, n uint64) string {
	ans := ""
	if n <= uint64(len(str)) {
		ans = str[n:]
	}
	return ans
}

var rootCmd = &cobra.Command{
	Use:     "uniq [-c | -d | -u] [-i] [-f num] [-s chars] [input_file [output_file]]",
	Short:   "UNIX uniq analog",
	Long:    "Filter adjacent matching lines from INPUT (or standard input),\nwriting to OUTPUT (or standard output).",
	Version: "1.0",
	Args:    cobra.MaximumNArgs(2),
	PreRun: func(cmd *cobra.Command, args []string) {
		ReadFlag := func(name string) int8 {
			flag, err := cmd.Flags().GetBool(name)
			if err != nil {
				os.Exit(1) // parse error
			}
			ans := int8(0)
			if flag {
				ans = 1
			}
			return ans
		}

		cFlag := ReadFlag("count")
		dFlag := ReadFlag("repeated")
		uFlag := ReadFlag("unique")

		if cFlag+dFlag+uFlag > 1 {
			os.Exit(1) // too many flags
		}

		var err error

		if len(args) >= 1 {
			inputStream, err = os.Open(args[0])
			if err != nil {
				os.Exit(1) // unable to open file
			}
		}
		if len(args) == 2 {
			outputStream, err = os.OpenFile(args[1], os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				os.Exit(1) // unable to open file
			}
		}
	},

	Run: func(cmd *cobra.Command, args []string) {

		ReadFlag := func(name string) int8 {
			flag, err := cmd.Flags().GetBool(name)
			if err != nil {
				os.Exit(1) // parse error
			}
			ans := int8(0)
			if flag {
				ans = 1
			}
			return ans
		}

		fmt.Println("args: ", args)
		scanner := bufio.NewScanner(inputStream)

		fFlagVal, _ := cmd.Flags().GetUint64("skip-fields")
		sFlagVal, _ := cmd.Flags().GetUint64("skip-chars")
		cFlag := ReadFlag("count")
		dFlag := ReadFlag("repeated")
		uFlag := ReadFlag("unique")

		currStrFormatted := ""
		prevStrFormatted := ""
		firstString := true
		for scanner.Scan() {
			currStr := scanner.Text()
			currStrFormatted = currStr
			if fFlagVal > 0 {
				currStrFormatted = TruncateNFields(currStrFormatted, fFlagVal)
			}
			if sFlagVal > 0 {
				currStrFormatted = TruncateNCharacters(currStrFormatted, sFlagVal)
			}
			if cFlag != 0 {
				PrintWithCountFlag(&prevStrFormatted, &currStrFormatted, firstString, false)
				if firstString {
					firstString = false
				}
			} else if dFlag != 0 {
				PrintWithRepeatedFlag(&prevStrFormatted, &currStrFormatted, false)
			} else if uFlag != 0 {
				PrintWithUniqueFlag(&prevStrFormatted, &currStrFormatted)
			} else {
				PrintWithNoFlag(&prevStrFormatted, &currStrFormatted)
			}
			prevStrFormatted = currStrFormatted
		}
		if dFlag != 0 {
			PrintWithRepeatedFlag(&prevStrFormatted, &currStrFormatted, true)
		}
		if cFlag != 0 {
			PrintWithCountFlag(&prevStrFormatted, &currStrFormatted, false, true)
			if firstString {
				firstString = false
			}
		}

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("count", "c", false, "prefix lines by the number of occurrences")
	rootCmd.Flags().BoolP("repeated", "d", false, "only print duplicate lines, one for each group")
	rootCmd.Flags().BoolP("unique", "u", false, "only print unique lines")
	rootCmd.Flags().BoolP("ignore-case", "i", false, "ignore differences in case when comparing")
	rootCmd.Flags().Uint64P("skip-fields", "f", 0, "avoid comparing the first N fields")
	rootCmd.Flags().Uint64P("skip-chars", "s", 0, "avoid comparing the first N characters")
}
