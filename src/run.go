package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"png-decoder/src/decode"
	"png-decoder/src/encode"
	"png-decoder/src/global"
)


var ( // errors
	errInvalidFlags = errors.New("invalid flags")
	errInvalidCmd = errors.New("invalid command")
	errInvalidSyntax = errors.New("invalid syntax")
)

var ( // regex patterns
	regActions = regexp.MustCompile("'\\^\\[\\[(A|B)|\\n$'m")
) // TODO: implement regex into os.Stdin reader


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
		head := strings.SplitN(cmd, " ", 2)
		switch {
		case strings.Compare("quit", head[0]) == 0:
			fmt.Println("--------------------------")
			break mainLoop
		case strings.Compare("help", head[0]) == 0:
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
		case strings.Compare("encode", head[0]) == 0:
			err = checkSyntax(head...)
			if err != nil {
				logError(err)
				continue mainLoop
			}
			err = encode.Encode()
			global.Reset()
			if err != nil {
				logError(err)
				continue mainLoop
			}
			log.Println("encoding successful")
		case strings.Compare("decode", head[0]) == 0:
			err = checkSyntax(head...)
			if err != nil {
				logError(err)
				continue mainLoop
			}
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
	log.Println("[ERROR]", err)
}

func checkSyntax(head... string) error {
	if len(head) == 1 {
		return errInvalidSyntax
	}
	if []rune("\"")[0] != rune(head[1][0]) {
		return errInvalidSyntax
	}
	path := strings.SplitAfter(head[1], "\" ")
	path[0] = strings.TrimSuffix(path[0], " ")
	if []rune("\"")[0] != rune(path[0][len(path[0])-1]) {
		return errInvalidSyntax
	}
	if len(path) > 1 {
		flags := strings.Split(path[1], " ")
		err := analyzeFlags(true, flags...)
		if err != nil {
			global.Reset()
			return err
		}
	}
	path[0] = strings.TrimPrefix(path[0], "\"")
	path[0] = strings.TrimSuffix(path[0], "\"")
	global.Path = path[0]
	return nil
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
				global.Inter = true
			case strings.Compare("false", parts[1]) == 0:
				global.Inter = false
			default:
				return errInvalidFlags
			}
		case strings.Compare("-alpha", parts[0]) == 0:
			switch {
			case strings.Compare("true", parts[1]) == 0:
				global.Alpha = true
			case strings.Compare("false", parts[1]) == 0:
				global.Alpha = false
			default:
				return errInvalidFlags
			}
		default:
			return errInvalidFlags
		}
	}
	return nil
}