package main

import "fmt"

// Request is a client's command/request
type Request struct {
	client    *Client
	reqString string
	args      []*RequestArg
}

// RequestArg - struct of a request argument
type RequestArg struct {
	Text      string
	RawText   string
	QuoteType int
}

// NewRequest returns an instantiated Request object
func NewRequest(client *Client, reqString string) *Request {
	request := &Request{}
	request.client = client
	request.reqString = reqString
	request.splitString()
	return request
}

func (r *Request) handle() {
	var cmd *Command
	var ok bool
	name := r.args[0].Text
	if name == "?" {
		name = "help"
	}

	if cmd, ok = commands[name]; !ok {
		outStr := fmt.Sprintf("invalid command: %s\n"+
			"Use \"help\" or \"?\" for help.\n",
			r.args[0].Text)
		r.client.send(outStr)
		return
	}
	cmd.run(r)
}

func (r *Request) splitString() {
	command := r.reqString
	var commArr []string
	var prevChar string
	var curChar string
	var nextArg string
	var nextArgRaw string
	inDoubleQuotes := false
	inSingleQuotes := false
	for len(command) > 0 {
		prevChar = curChar
		curChar = command[0:1]
		command = command[1:]

		switch curChar {
		case " ":
			if !inDoubleQuotes && !inSingleQuotes && prevChar != "\\" {
				// separator
				if len(nextArg) > 0 {
					commArr = append(commArr, nextArg)
					r.args = append(r.args, &RequestArg{
						Text:    nextArg,
						RawText: nextArgRaw,
					})
					nextArg = ""
					nextArgRaw = ""
				}
			} else {
				nextArg += curChar
				nextArgRaw += curChar
			}
		case "\"":
			nextArgRaw += curChar
			if prevChar != "\\" && !inSingleQuotes {
				inDoubleQuotes = !inDoubleQuotes
			} else {
				nextArg += curChar
			}
		case "'":
			nextArgRaw += curChar
			if prevChar != "\\" && !inDoubleQuotes {
				inSingleQuotes = !inSingleQuotes
			} else {
				nextArg = nextArg + curChar
			}
		default:
			nextArg += curChar
			nextArgRaw += curChar
		}
	}
	if len(nextArgRaw) > 0 {
		r.args = append(r.args, &RequestArg{
			Text:    nextArg,
			RawText: nextArgRaw,
		})
	}
}
