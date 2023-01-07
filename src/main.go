package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"png-decoder/src/decode"
	"png-decoder/src/encode"
	"png-decoder/src/global"
)


var (
	errInvalidFlags = errors.New("invalid flags")
	errInvalidCmd = errors.New("invalid command")
)


func main() {
	r := bufio.NewReader(os.Stdin)
	fmt.Println("png encoder/decoder")
	fmt.Println("--------------------------")
	fmt.Println("type help for help")
	mainLoop: for {
		fmt.Print("> ")
		cmd, err := r.ReadString('\n')
		if err != nil {
			logError(err)
			continue mainLoop
		}
		cmd = strings.TrimSuffix(cmd, "\n")
		tokens := strings.Split(cmd, " ")
		switch {
		case strings.Compare("quit", tokens[0]) == 0:
			fmt.Println("--------------------------")
			break mainLoop
		case strings.Compare("help", tokens[0]) == 0:
			fmt.Println("encode: encodes BMP as PNG")
			fmt.Println("syntax: encode \"path\"")
			fmt.Println("flags: -alpha=TRUE/false (enable alpha channel), -inter=true/FALSE (enable adam7 interlacing)")
			fmt.Println()
			fmt.Println("decode: decodes PNG to BMP")
			fmt.Println("syntax: decode \"path\"")
			fmt.Println("flags: -alpha=TRUE/false (enables alpha channel)")
			fmt.Println()
			fmt.Println("help: prints help for available commands")
			fmt.Println()
			fmt.Println("quit: exit the application")
			continue mainLoop
		case strings.Compare("encode", tokens[0]) == 0:
			err = analyzeFlags(true, tokens[2:]...)
			if err != nil {
				global.Reset()
				logError(err)
				continue mainLoop
			}
			tokens[1] = strings.TrimPrefix(tokens[1], "\"")
			tokens[1] = strings.TrimSuffix(tokens[1], "\"")
			global.SetPath(tokens[1])
			err = encode.Encode()
			global.Reset()
			if err != nil {
				logError(err)
				continue mainLoop
			}
			log.Println("encoding successful")
		case strings.Compare("decode", tokens[0]) == 0:
			err = analyzeFlags(false, tokens[2:]...)
			if err != nil {
				global.Reset()
				logError(err)
				continue mainLoop
			}
			tokens[1] = strings.TrimPrefix(tokens[1], "\"")
			tokens[1] = strings.TrimSuffix(tokens[1], "\"")
			global.SetPath(tokens[1])
			err = decode.Decode()
			global.Reset()
			if err != nil {
				logError(err)
				continue mainLoop
			}
			log.Println("decoding successful")
		default:
			logError(errInvalidCmd)
			continue mainLoop
		}
	}
}

func logError(err error) {
	if err == decode.WarnUnknownAncChunk {
		log.Println("[WARNING]", err)
	} else {
		log.Println("[ERROR]", err)
	}
}

func analyzeFlags(enc bool, flags... string) error {
	for _, f := range flags {
		parts := strings.Split(f, "=")
		if len(parts) != 2 {
			return errInvalidFlags
		}
		switch {
		case strings.Compare("-inter", parts[0]) == 0 && enc:
			switch {
			case strings.Compare("true", parts[1]) == 0:
				global.SetInterlaced(true)
			case strings.Compare("false", parts[1]) == 0:
				global.SetInterlaced(false)
			default:
				return errInvalidFlags
			}
		case strings.Compare("-alpha", parts[0]) == 0:
			switch {
			case strings.Compare("true", parts[1]) == 0:
				global.SetAlpha(true)
			case strings.Compare("false", parts[1]) == 0:
				global.SetAlpha(false)
			default:
				return errInvalidFlags
			}
		default:
			return errInvalidFlags
		}
	}
	return nil
}