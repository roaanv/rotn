// Command rotn applies a rot-N byte shift to arbitrary (binary) data.
//
// Every byte b is encoded as (b + n) mod 256 and decoded as (b - n) mod 256,
// so decode(encode(x, n), n) == x for any integer n.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `Usage:
  rotn encode [-n N] [-f INPUT] [-o OUTPUT]
  rotn decode [-n N] [-f INPUT] [-o OUTPUT]

Shifts every byte by N (mod 256). Decode reverses encode with the same N.
N may be any integer (negative or > 255); it wraps mod 256. Default 13.
INPUT defaults to stdin, OUTPUT to stdout.
`

const bufSize = 64 * 1024

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	cmd := args[0]
	switch cmd {
	case "encode", "decode":
	case "-h", "-help", "--help", "help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "rotn: unknown command %q\n\n%s", cmd, usage)
		return 2
	}

	fs := flag.NewFlagSet("rotn "+cmd, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, usage) }
	n := fs.Int("n", 13, "rotation amount (any integer, wraps mod 256)")
	in := fs.String("f", "", "input file (default stdin)")
	out := fs.String("o", "", "output file (default stdout)")
	if err := fs.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "rotn: unexpected arguments: %v\n", fs.Args())
		return 2
	}

	k := key(*n)
	if cmd == "decode" {
		k = -k // byte arithmetic: additive inverse mod 256
	}

	if err := process(*in, *out, k, stdin, stdout); err != nil {
		fmt.Fprintf(stderr, "rotn: %v\n", err)
		return 1
	}
	return 0
}

// key reduces any integer n to its equivalent shift in [0, 255].
func key(n int) byte {
	m := n % 256
	if m < 0 {
		m += 256
	}
	return byte(m)
}

// shift adds k to every byte of buf in place, wrapping mod 256.
func shift(buf []byte, k byte) {
	for i := range buf {
		buf[i] += k
	}
}

func process(inPath, outPath string, k byte, stdin io.Reader, stdout io.Writer) (err error) {
	var r io.Reader = stdin
	var inInfo os.FileInfo
	if inPath != "" {
		f, err := os.Open(inPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if inInfo, err = f.Stat(); err != nil {
			return err
		}
		r = f
	}

	w := stdout
	if outPath != "" {
		// Creating the output truncates it; if it is the input file, the data
		// would be destroyed before it is read.
		if inInfo != nil {
			if outInfo, statErr := os.Stat(outPath); statErr == nil && os.SameFile(inInfo, outInfo) {
				return fmt.Errorf("input and output are the same file: %s", outPath)
			}
		}
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); err == nil {
				err = cerr
			}
		}()
		w = f
	}

	return copyShifted(w, r, k)
}

func copyShifted(w io.Writer, r io.Reader, k byte) error {
	buf := make([]byte, bufSize)
	for {
		nr, rerr := r.Read(buf)
		if nr > 0 {
			chunk := buf[:nr]
			shift(chunk, k)
			if _, werr := w.Write(chunk); werr != nil {
				return werr
			}
		}
		if rerr == io.EOF {
			return nil
		}
		if rerr != nil {
			return rerr
		}
	}
}
