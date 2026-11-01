# rotn

A rot13-style tool for binary files that works with any rotation amount. Every byte `b` becomes `(b + n) mod 256` when encoding and `(b - n) mod 256` when decoding.

## Install

```sh
make build      # builds bin/rotn
make install    # builds and copies to $HOME/.local/bin/rotn
make install BINDIR=/usr/local/bin   # custom location
make uninstall  # removes the installed binary
make test
make clean
```

## Usage

```sh
rotn encode -n 13 -f input.bin -o encoded.bin
rotn decode -n 13 -f encoded.bin -o decoded.bin
```

| Flag | Default | Meaning |
|------|---------|---------|
| `-n` | `13` | Rotation amount. Any integer works and wraps mod 256, so `-3` is the same as `253`. |
| `-f` | stdin | Input file |
| `-o` | stdout | Output file; an existing file is overwritten |

The tool streams data, so file size is not limited by memory. It can be used in a pipe:

```sh
cat photo.jpg | rotn encode -n 42 | rotn decode -n 42 > same.jpg
```

Using the same path for `-f` and `-o` is rejected, because opening the output would truncate the input.

`n = 128` reverses itself: encode and decode give the same result, as rot13 does for letters.
