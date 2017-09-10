# hambot
## Setup
`hambot` reads its configuration from a JSON file. It is read from, in order:
1. The path supplied as an argument to `hambot`
2. The file `config.json` in the same directory as the `hambot` executable
3. The file `hambot/config.json` in one of the standard configuration paths
   (see [https://github.com/shibukawa/configdir]).

See [main.go] for a description of the format.

## Submission guidelines
Format source files with `gofmt` and fix issues reported by `golint`.
These commands are executed automatically by some editors,
such as [Visual Studio Code](http://code.visualstudio.com/).