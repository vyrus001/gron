package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/nwidger/jsoncolor"
	"github.com/pkg/errors"
)

// Exit codes
const (
	exitOK = iota
	exitOpenFile
	exitReadInput
	exitFormStatements
	exitFetchURL
	exitParseStatements
	exitJSONEncode
)

// Option bitfields
const (
	optMonochrome = 1 << iota
	optNoSort
	optJSON
	optStream
)

// Output colors
var (
	strColor   = color.New(color.FgYellow)
	braceColor = color.New(color.FgMagenta)
	bareColor  = color.New(color.FgBlue, color.Bold)
	numColor   = color.New(color.FgRed)
	boolColor  = color.New(color.FgCyan)
)

// gronVersion stores the current gron version, set at build
// time with the ldflags -X option
var gronVersion = "dev"

// Default value that is set if proxy/noproxy variables is not specified.
// Cannot be an empty string, because an empty string explicitly deactivates the
// proxy.
var undefinedProxy = "-"

func init() {
	flag.Usage = func() {
		h := "Transform JSON (from a file, URL, or stdin) into discrete assignments to make it greppable\n\n"

		h += "Usage:\n"
		h += "  gron [OPTIONS] [FILE|URL|-]\n\n"

		h += "Options:\n"
		h += "  -u, --ungron     Reverse the operation (turn assignments back into JSON)\n"
		h += "  -v, --values     Print just the values of provided assignments\n"
		h += "  -c, --colorize, --color Colorize output (default on tty)\n"
		h += "  -m, --monochrome, --no-color Monochrome output\n"
		h += "  -s, --stream     Process consecutive JSON values as an array\n"
		h += "  -k, --insecure   Disable certificate validation\n"
		h += "  -x, --proxy      Set proxy configuration\n"
		h += "      --noproxy    Comma-separated list of hosts for which not to use a proxy, if one is specified.\n"
		h += "  -j, --json       Represent gron data as JSON stream\n"
		h += "  -H, --with-filename Prefix text output lines with the input name\n"
		h += "      --no-sort    Stream output in input order with lower memory use\n"
		h += "      --version    Print version information\n\n"

		h += "Exit Codes:\n"
		h += fmt.Sprintf("  %d\t%s\n", exitOK, "OK")
		h += fmt.Sprintf("  %d\t%s\n", exitOpenFile, "Failed to open file")
		h += fmt.Sprintf("  %d\t%s\n", exitReadInput, "Failed to read input")
		h += fmt.Sprintf("  %d\t%s\n", exitFormStatements, "Failed to form statements")
		h += fmt.Sprintf("  %d\t%s\n", exitFetchURL, "Failed to fetch URL")
		h += fmt.Sprintf("  %d\t%s\n", exitParseStatements, "Failed to parse statements")
		h += fmt.Sprintf("  %d\t%s\n", exitJSONEncode, "Failed to encode JSON")
		h += "\n"

		h += "Examples:\n"
		h += "  gron /tmp/apiresponse.json\n"
		h += "  gron http://jsonplaceholder.typicode.com/users/1 \n"
		h += "  curl -s http://jsonplaceholder.typicode.com/users/1 | gron\n"
		h += "  gron http://jsonplaceholder.typicode.com/users/1 | grep company | gron --ungron\n"

		fmt.Fprint(os.Stderr, h)
	}
}

func main() {
	var (
		ungronFlag     bool
		colorizeFlag   bool
		monochromeFlag bool
		streamFlag     bool
		noSortFlag     bool
		versionFlag    bool
		insecureFlag   bool
		jsonFlag       bool
		valuesFlag     bool
		filenameFlag   bool
		proxyURL       string
		noProxy        string
	)

	flag.BoolVar(&ungronFlag, "ungron", false, "")
	flag.BoolVar(&ungronFlag, "u", false, "")
	flag.BoolVar(&colorizeFlag, "colorize", false, "")
	flag.BoolVar(&colorizeFlag, "c", false, "")
	flag.BoolVar(&colorizeFlag, "color", false, "")
	flag.BoolVar(&monochromeFlag, "monochrome", false, "")
	flag.BoolVar(&monochromeFlag, "m", false, "")
	flag.BoolVar(&monochromeFlag, "no-color", false, "")
	flag.BoolVar(&streamFlag, "s", false, "")
	flag.BoolVar(&streamFlag, "stream", false, "")
	flag.BoolVar(&noSortFlag, "no-sort", false, "")
	flag.BoolVar(&versionFlag, "version", false, "")
	flag.BoolVar(&insecureFlag, "k", false, "")
	flag.BoolVar(&insecureFlag, "insecure", false, "")
	flag.BoolVar(&jsonFlag, "j", false, "")
	flag.BoolVar(&jsonFlag, "json", false, "")
	flag.BoolVar(&valuesFlag, "values", false, "")
	flag.BoolVar(&valuesFlag, "value", false, "")
	flag.BoolVar(&valuesFlag, "v", false, "")
	flag.BoolVar(&filenameFlag, "H", false, "")
	flag.BoolVar(&filenameFlag, "with-filename", false, "")
	flag.StringVar(&proxyURL, "x", undefinedProxy, "")
	flag.StringVar(&proxyURL, "proxy", undefinedProxy, "")
	flag.StringVar(&noProxy, "noproxy", undefinedProxy, "")

	flag.Parse()

	// Print version information
	if versionFlag {
		fmt.Printf("gron version %s\n", version())
		os.Exit(exitOK)
	}

	// If executed as 'ungron' set the --ungron flag
	if strings.HasSuffix(os.Args[0], "ungron") {
		ungronFlag = true
	}

	// Determine what the program's input should be:
	// file, HTTP URL or stdin
	var rawInput io.Reader
	var inputCloser io.Closer
	filename := flag.Arg(0)
	if filename == "" || filename == "-" {
		rawInput = os.Stdin
	} else if validURL(filename) {
		r, err := getURL(filename, insecureFlag, proxyURL, noProxy)
		if err != nil {
			fatal(exitFetchURL, err)
		}
		rawInput = r
		inputCloser = r
	} else {
		r, err := os.Open(filename)
		if err != nil {
			fatal(exitOpenFile, err)
		}
		rawInput = r
		inputCloser = r
	}

	var opts int
	// The monochrome option should be forced if the output isn't a terminal
	// to avoid doing unnecessary work calling the color functions
	switch {
	case colorizeFlag:
		color.NoColor = false
	case monochromeFlag || color.NoColor:
		opts = opts | optMonochrome
	}
	if noSortFlag {
		opts = opts | optNoSort
	}
	if jsonFlag {
		opts = opts | optJSON
	}
	if streamFlag {
		opts = opts | optStream
		if ungronFlag {
			opts = opts | optMonochrome
		}
	}
	if filenameFlag && (jsonFlag || ungronFlag) {
		fatal(exitParseStatements, fmt.Errorf("--with-filename requires text gron output"))
	}

	// Pick the appropriate action: gron, ungron, gronValues, or gronStream
	var a actionFn = gron
	if ungronFlag {
		a = ungron
	} else if valuesFlag {
		a = gronValues
	} else if streamFlag {
		a = gronStream
	}
	var output io.Writer = colorable.NewColorableStdout()
	if filenameFlag {
		if filename == "" {
			filename = "-"
		}
		output = newLinePrefixWriter(output, filename+": ")
	}
	exitCode, err := a(rawInput, output, opts)
	if inputCloser != nil {
		_ = inputCloser.Close()
	}

	if exitCode != exitOK {
		fatal(exitCode, err)
	}

	os.Exit(exitOK)
}

func version() string {
	if gronVersion != "dev" {
		return gronVersion
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return gronVersion
}

// an actionFn represents a main action of the program, it accepts
// an input, output and a bitfield of options; returning an exit
// code and any error that occurred
type actionFn func(io.Reader, io.Writer, int) (int, error)

// gron is the default action. Given JSON as the input it returns a list
// of assignment statements. Possible options are optNoSort and optMonochrome
func gron(r io.Reader, w io.Writer, opts int) (int, error) {
	if opts&optNoSort > 0 {
		return gronUnsorted(r, w, opts)
	}
	var err error

	// Buffer the output; one write per statement is slow. Lines are
	// never split between writes, so color sequences stay whole.
	bw := bufio.NewWriter(w)

	var conv statementconv
	if opts&optMonochrome > 0 {
		conv = statementToString
	} else {
		conv = statementToColorString
	}

	ss, err := statementsFromJSON(r, statement{{"json", typBare}})
	if err != nil {
		goto out
	}

	// Go's maps do not have well-defined ordering, but we want a consistent
	// output for a given input, so we must sort the statements
	if opts&optNoSort == 0 {
		sort.Sort(ss)
	}

	for _, s := range ss {
		if opts&optJSON > 0 {
			s, err = s.jsonify()
			if err != nil {
				goto out
			}
		}
		line := conv(s)
		if bw.Available() < len(line)+1 {
			if flushErr := bw.Flush(); flushErr != nil {
				return exitJSONEncode, fmt.Errorf("failed to write output: %w", flushErr)
			}
		}
		if _, writeErr := fmt.Fprintln(bw, line); writeErr != nil {
			return exitJSONEncode, fmt.Errorf("failed to write output: %w", writeErr)
		}
	}

out:
	if err != nil {
		return exitFormStatements, fmt.Errorf("failed to form statements: %s", err)
	}
	if err := bw.Flush(); err != nil {
		return exitJSONEncode, fmt.Errorf("failed to write output: %w", err)
	}
	return exitOK, nil
}

// gronStream is like the gron action, but processes consecutive JSON values.
func gronStream(r io.Reader, w io.Writer, opts int) (int, error) {
	var conv func(s statement) string
	if opts&optMonochrome > 0 {
		conv = statementToString
	} else {
		conv = statementToColorString
	}

	// Helper function to make the prefix statements for each line
	makePrefix := func(index int) statement {
		return statement{
			{"json", typBare},
			{"[", typLBrace},
			{fmt.Sprintf("%d", index), typNumericKey},
			{"]", typRBrace},
		}
	}

	// The first line of output needs to establish that the top-level
	// thing is actually an array...
	top := statement{
		{"json", typBare},
		{"=", typEquals},
		{"[]", typEmptyArray},
		{";", typSemi},
	}

	if opts&optJSON > 0 {
		var err error
		top, err = top.jsonify()
		if err != nil {
			return exitFormStatements, fmt.Errorf("failed to form statements: %w", err)
		}
	}

	fmt.Fprintln(w, conv(top))

	decoder := json.NewDecoder(withoutUTF8BOM(r))
	decoder.UseNumber()
	for i := 0; ; i++ {
		if opts&optNoSort > 0 {
			first, err := decoder.Token()
			if err == io.EOF {
				return exitOK, nil
			}
			if err != nil {
				return exitFormStatements, fmt.Errorf("failed to decode JSON value %d: %w", i+1, err)
			}
			err = walkJSON(decoder, first, makePrefix(i), func(s statement) error {
				if opts&optJSON > 0 {
					var err error
					s, err = s.jsonify()
					if err != nil {
						return err
					}
				}
				_, err := fmt.Fprintln(w, conv(s))
				return err
			})
			if err != nil {
				return exitFormStatements, fmt.Errorf("failed to decode JSON value %d: %w", i+1, err)
			}
			continue
		}
		var value interface{}
		err := decoder.Decode(&value)
		if err == io.EOF {
			return exitOK, nil
		}
		if err != nil {
			return exitFormStatements, fmt.Errorf("failed to decode JSON value %d: %w", i+1, err)
		}
		ss := make(statements, 0, 32)
		ss.fill(makePrefix(i), value)

		// Go's maps do not have well-defined ordering, but we want a consistent
		// output for a given input, so we must sort the statements
		if opts&optNoSort == 0 {
			sort.Sort(ss)
		}

		for _, s := range ss {
			if opts&optJSON > 0 {
				s, err = s.jsonify()
				if err != nil {
					return exitFormStatements, fmt.Errorf("failed to form statements: %w", err)
				}
			}
			fmt.Fprintln(w, conv(s))
		}
	}
}

// ungron is the reverse of gron. Given assignment statements as input,
// it returns JSON. The only option is optMonochrome
func ungron(r io.Reader, w io.Writer, opts int) (int, error) {
	reader := bufio.NewReader(r)
	var maker statementmaker

	if opts&optJSON > 0 {
		maker = statementFromJSONSpec
	} else {
		maker = statementFromStringMaker
	}

	// Parse and merge each statement before reading the next one.
	var merger statementMerger
	for lineNo := 1; ; lineNo++ {
		line, readErr := reader.ReadString('\n')
		if readErr == io.EOF && line == "" {
			break
		}
		if readErr != nil && readErr != io.EOF {
			return exitReadInput, fmt.Errorf("failed to read input statements at line %d: %w", lineNo, readErr)
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		s, err := maker(line)
		if err != nil {
			return exitParseStatements, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if err := merger.add(s); err != nil {
			return exitParseStatements, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if readErr == io.EOF {
			break
		}
	}

	if merger.count == 0 {
		return exitParseStatements, fmt.Errorf("no statements were parsed")
	}
	merged := merger.value

	// If there's only one top level key and it's "json", make that the top level thing
	mergedMap, ok := merged.(map[string]interface{})
	if ok {
		if len(mergedMap) == 1 {
			if _, exists := mergedMap["json"]; exists {
				merged = mergedMap["json"]
			}
		}
	}
	if opts&optStream > 0 {
		values, ok := merged.([]interface{})
		if !ok {
			return exitParseStatements, fmt.Errorf("streamed ungron input must be an array")
		}
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		for _, value := range values {
			if err := enc.Encode(value); err != nil {
				return exitJSONEncode, errors.Wrap(err, "failed to convert statements to JSON")
			}
		}
		return exitOK, nil
	}

	if opts&optMonochrome > 0 {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(merged); err != nil {
			return exitJSONEncode, errors.Wrap(err, "failed to convert statements to JSON")
		}
		return exitOK, nil
	}

	// Colorized output needs the full JSON document.
	out := &bytes.Buffer{}
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	err := enc.Encode(merged)
	if err != nil {
		return exitJSONEncode, errors.Wrap(err, "failed to convert statements to JSON")
	}
	j := out.Bytes()

	// If the output isn't monochrome, add color to the JSON
	if opts&optMonochrome == 0 {
		c, err := colorizeJSON(j)

		// If we failed to colorize the JSON for whatever reason,
		// we'll just fall back to monochrome output, otherwise
		// replace the monochrome JSON with glorious technicolor
		if err == nil {
			j = c
		}
	}

	// For whatever reason, the monochrome version of the JSON
	// has a trailing newline character, but the colorized version
	// does not. Strip the whitespace so that neither has the newline
	// character on the end, and then we'll add a newline in the
	// Fprintf below
	j = bytes.TrimSpace(j)

	fmt.Fprintf(w, "%s\n", j)

	return exitOK, nil
}

// gronValues prints just the scalar values from some input gron statements
// without any quotes or anything of that sort; a bit like jq -r
// e.g. json[0].user.name = "Sam"; -> Sam
func gronValues(r io.Reader, w io.Writer, opts int) (int, error) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		s := statementFromString(scanner.Text())

		if len(s) == 0 {
			return exitParseStatements, fmt.Errorf("failed to parse '%s' as gron statement", scanner.Text())
		}

		// strip off the leading 'json' bare key
		if s[0].typ == typBare && s[0].text == "json" {
			s = s[1:]
		}

		// strip off the leading dots
		if len(s) > 0 && (s[0].typ == typDot || s[0].typ == typLBrace) {
			s = s[1:]
		}

		for _, t := range s {
			switch t.typ {
			case typString:
				var text string
				err := json.Unmarshal([]byte(t.text), &text)
				if err != nil {
					// just swallow errors and try to continue
					continue
				}
				fmt.Fprintln(w, text)

			case typNumber, typTrue, typFalse, typNull:
				fmt.Fprintln(w, t.text)

			default:
				// Nothing
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return exitReadInput, err
	}

	return exitOK, nil
}

func colorizeJSON(src []byte) ([]byte, error) {
	out := &bytes.Buffer{}
	f := jsoncolor.NewFormatter()

	f.StringColor = strColor
	f.ObjectColor = braceColor
	f.ArrayColor = braceColor
	f.FieldColor = bareColor
	f.NumberColor = numColor
	f.TrueColor = boolColor
	f.FalseColor = boolColor
	f.NullColor = boolColor

	err := f.Format(out, src)
	if err != nil {
		return out.Bytes(), err
	}
	return out.Bytes(), nil
}

func fatal(code int, err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(code)
}
