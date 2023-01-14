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


const ( // help command output
	helpF = "encode: encodes BMP as PNG\n"+
			"syntax: encode \"path\" -flags\n"+
			"flags: -alpha=true/false (enable alpha channel), -inter=true/false (enable adam7 interlacing)\n"+
			"\n"+
			"decode: decodes PNG to BMP\n"+
			"syntax: decode \"path\" -flags\n"+
			"flags: -alpha=true/false (enables alpha channel)\n"+
			"\n"+
			"help: prints help for available commands\n"+
			"\n"+
			"quit: exit the application\n"
)

var ( // errors
	errInvalidFlags = errors.New("invalid flags")
	errInvalidCmd = errors.New("invalid command")
	errInvalidSyntax = errors.New("invalid syntax")
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
		head := strings.SplitN(cmd, " ", 2)
		switch {
		case strings.Compare("quit", head[0]) == 0:
			fmt.Println("--------------------------")
			break mainLoop
		case strings.Compare("help", head[0]) == 0:
			fmt.Printf(helpF)
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