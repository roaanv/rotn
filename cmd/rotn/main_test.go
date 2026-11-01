package main

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func allBytes() []byte {
	b := make([]byte, 256)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

func runCLI(t *testing.T, stdin []byte, args ...string) (stdout []byte, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(args, bytes.NewReader(stdin), &out, &errOut)
	if code != 0 {
		t.Logf("stderr: %s", errOut.String())
	}
	return out.Bytes(), code
}

func TestRoundTripAllBytes(t *testing.T) {
	for _, n := range []int{0, 1, 13, 128, 255, 256, 1000, -1, -13, -300, math.MaxInt, math.MinInt} {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			ns := strconv.Itoa(n)
			enc, code := runCLI(t, allBytes(), "encode", "-n", ns)
			if code != 0 {
				t.Fatalf("encode exit %d", code)
			}
			dec, code := runCLI(t, enc, "decode", "-n", ns)
			if code != 0 {
				t.Fatalf("decode exit %d", code)
			}
			if !bytes.Equal(dec, allBytes()) {
				t.Fatalf("round trip mismatch for n=%d", n)
			}
		})
	}
}

func TestEncodeWrapsMod256(t *testing.T) {
	in := []byte{'a', 'z', 0x00, 0xF5, 0xFF}
	want := []byte{'n', 'z' + 13, 0x0D, 0x02, 0x0C}
	for _, n := range []string{"13", "269", "-243"} {
		got, code := runCLI(t, in, "encode", "-n", n)
		if code != 0 || !bytes.Equal(got, want) {
			t.Fatalf("n=%s: got %v (exit %d), want %v", n, got, code, want)
		}
	}
}

func TestDefaultNIs13(t *testing.T) {
	got, _ := runCLI(t, []byte("a"), "encode")
	if string(got) != "n" {
		t.Fatalf("got %q, want %q", got, "n")
	}
}

func TestFileRoundTripLargerThanBuffer(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "in.bin")
	enc := filepath.Join(dir, "enc.bin")
	dec := filepath.Join(dir, "dec.bin")

	data := bytes.Repeat(allBytes(), bufSize/256*3+7) // spans several reads
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, code := runCLI(t, nil, "encode", "-n", "77", "-f", src, "-o", enc); code != 0 {
		t.Fatalf("encode exit %d", code)
	}
	encoded, _ := os.ReadFile(enc)
	if bytes.Equal(encoded, data) {
		t.Fatal("encoded output equals input")
	}
	if _, code := runCLI(t, nil, "decode", "-n", "77", "-f", enc, "-o", dec); code != 0 {
		t.Fatalf("decode exit %d", code)
	}
	got, _ := os.ReadFile(dec)
	if !bytes.Equal(got, data) {
		t.Fatal("decoded file differs from original")
	}
}

func TestSameInputOutputRefusedAndPreserved(t *testing.T) {
	src := filepath.Join(t.TempDir(), "f.bin")
	if err := os.WriteFile(src, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, code := runCLI(t, nil, "encode", "-f", src, "-o", src); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	got, _ := os.ReadFile(src)
	if string(got) != "keep me" {
		t.Fatalf("input modified: %q", got)
	}
}

func TestBadInvocations(t *testing.T) {
	cases := map[string][]string{
		"no command":      {},
		"unknown command": {"rotate"},
		"bad n":           {"encode", "-n", "abc"},
		"stray arg":       {"encode", "extra"},
	}
	for name, args := range cases {
		if _, code := runCLI(t, nil, args...); code != 2 {
			t.Errorf("%s: exit %d, want 2", name, code)
		}
	}
	if _, code := runCLI(t, nil, "encode", "-f", filepath.Join(t.TempDir(), "missing")); code != 1 {
		t.Errorf("missing input: exit %d, want 1", code)
	}
}
